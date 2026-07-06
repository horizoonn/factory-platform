package health

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"google.golang.org/grpc"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"

	postgrespool "github.com/horizoonn/factory-platform/platform/pkg/database/postgres/pool"
)

type Checker struct {
	pool         postgrespool.Pool
	grpcConns    map[string]*grpc.ClientConn
	checkTimeout time.Duration
}

func NewChecker(pool postgrespool.Pool, grpcConns map[string]*grpc.ClientConn) *Checker {
	return &Checker{
		pool:         pool,
		grpcConns:    grpcConns,
		checkTimeout: 3 * time.Second,
	}
}

func (c *Checker) Handler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), c.checkTimeout)
	defer cancel()

	checks := make(map[string]string)
	healthy := true

	if err := c.pool.Ping(ctx); err != nil {
		checks["database"] = err.Error()
		healthy = false
	} else {
		checks["database"] = "ok"
	}

	for name, conn := range c.grpcConns {
		if err := c.checkGRPC(ctx, conn); err != nil {
			checks[name] = err.Error()
			healthy = false
		} else {
			checks[name] = "ok"
		}
	}

	w.Header().Set("Content-Type", "application/json")
	if !healthy {
		w.WriteHeader(http.StatusServiceUnavailable)
	} else {
		w.WriteHeader(http.StatusOK)
	}
	if err := json.NewEncoder(w).Encode(checks); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
	}
}

func (c *Checker) checkGRPC(ctx context.Context, conn *grpc.ClientConn) error {
	client := healthpb.NewHealthClient(conn)
	resp, err := client.Check(ctx, &healthpb.HealthCheckRequest{Service: ""})
	if err != nil {
		return err
	}
	if resp.GetStatus() != healthpb.HealthCheckResponse_SERVING {
		return context.DeadlineExceeded
	}
	return nil
}
