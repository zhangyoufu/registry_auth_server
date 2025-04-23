package main

import (
	"context"
	"flag"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"time"

	yaml "github.com/goccy/go-yaml"
	ladon "github.com/ory/ladon"
	lagon_manager_memory "github.com/ory/ladon/manager/memory"
	registry_auth "github.com/zhangyoufu/registry_auth"
)

func main() {
	var (
		cloudflare     bool
		listenNetwork  string
		listenAddress  string
		services       = map[string]struct{}{}
		keyPath        string
		tokenIssuer    string
		tokenExpire    time.Duration
		credentialPath string
		policyPath     string
	)
	flag.BoolVar(&cloudflare, "cloudflare", false, "trust CF-Connecting-IP HTTP request header")
	flag.StringVar(&listenNetwork, "network", "tcp", "listening network")
	flag.StringVar(&listenAddress, "address", "0.0.0.0:8000", "listening address")
	flag.Func("service", "service/audience allowlist", func(s string) error { services[s] = struct{}{}; return nil })
	flag.StringVar(&keyPath, "key", "/run/secrets/auth_privkey", "path to PEM format private key")
	flag.StringVar(&tokenIssuer, "issuer", "registry-token-issuer", "token issuer")
	flag.DurationVar(&tokenExpire, "expire", 5*time.Minute, "token expiration time")
	flag.StringVar(&credentialPath, "cred", "/mnt/passwd", "path to htpasswd format credential file")
	flag.StringVar(&policyPath, "policy", "/mnt/policy", "path to policy folder")
	flag.Parse()

	// check services
	if len(services) == 0 {
		log.Fatal("service is not specified")
	}

	// load credentials
	db, err := loadPasswordDatabase(credentialPath)
	if err != nil {
		log.Fatalf("failed to load credentials: %v", err)
	}

	// load policies
	policyManager := lagon_manager_memory.NewMemoryManager()
	policyEntries, err := os.ReadDir(policyPath)
	if err != nil {
		log.Fatalf("failed to load policies: %v", err)
	}
	for _, entry := range policyEntries {
		ext := filepath.Ext(entry.Name())
		if ext != ".yml" && ext != ".yaml" {
			continue
		}
		policyBytes, err := os.ReadFile(filepath.Join(policyPath, entry.Name()))
		if err != nil {
			log.Printf("failed to load policy file %s: %v", entry.Name(), err)
			continue
		}
		policy := ladon.DefaultPolicy{}
		err = yaml.UnmarshalWithOptions(policyBytes, &policy, yaml.Strict(), yaml.UseJSONUnmarshaler())
		if err != nil {
			log.Printf("failed to load policy file %s: %v", entry.Name(), err)
			continue
		}
		err = policyManager.Create(context.Background(), &policy)
		if err != nil {
			log.Printf("failed to load policy file %s: %v", entry.Name(), err)
			continue
		}
		log.Printf("loaded policy file %s", entry.Name())
	}

	// load private key
	tokenSigner, err := registry_auth.LoadPrivateKey(keyPath)
	if err != nil {
		log.Fatalf("failed to load private key: %v", err)
	}

	// initialize registry_auth handler
	authHandler, err := registry_auth.NewHandler(registry_auth.HandlerConfig{
		Authenticator: &Authenticator{db: db},
		Authorizer:    &Authorizer{warden: &ladon.Ladon{Manager: policyManager}},
		Services:      services,
		TokenSigner:   tokenSigner,
		TokenIssuer:   tokenIssuer,
		TokenExpire:   tokenExpire,
	})
	if err != nil {
		log.Fatalf("failed to initialize auth handler: %v", err)
	}

	// initialize HTTP request dispatching
	httpMux := http.NewServeMux()
	httpMux.HandleFunc("GET /token", authHandler.HandleToken)
	httpMux.HandleFunc("POST /token", authHandler.HandleOAuth2Token)
	httpHandler := http.Handler(httpMux)

	// install Cloudflare RemoteAddr middleware
	if cloudflare {
		httpHandler = cloudflareRemoteAddrMiddleware(httpHandler)
	}

	// listen
	ln, err := net.Listen(listenNetwork, listenAddress)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("listening on %s %s", listenNetwork, listenAddress)

	// serve
	log.Fatal(http.Serve(ln, httpHandler))
}
