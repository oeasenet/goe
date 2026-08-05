// Package mongodb provides the optional MongoDB module. It exposes the
// contract.MongoDB abstraction and wires lifecycle hooks to manage a single
// client connection safely through Fx. Applications needing a second data
// source construct their own mongo.Client via the driver.
//
// # Configuration
//
// Configure via environment variables:
//
//	MONGO_URI=mongodb://localhost:27017  # Connection URI (environment-only)
//	MONGO_DB_NAME=myapp                  # Database name
//	MONGO_USERNAME=svc                   # Username (environment-only)
//	MONGO_PASSWORD=secret                # Password (environment-only)
//	MONGO_MAX_POOL_SIZE=100              # Connection pool ceiling
//	MONGO_PING_TIMEOUT=5s                # Startup reachability ping timeout
//
// Or from Go code through goe.Options.MongoDB — every MONGO_* key has a
// matching Option (MONGO_MAX_POOL_SIZE is mongodb.WithMaxPoolSize;
// connection-scoped keys take the connection name first), and code wins over
// the environment:
//
//	goe.New(goe.Options{
//	    MongoDB: []mongodb.Option{
//	        mongodb.WithDatabase("myapp"),
//	        mongodb.WithMaxPoolSize(50),
//	    },
//	})
//
// Credentials are the exception: MONGO_URI, MONGO_USERNAME and MONGO_PASSWORD
// have no Option. Secrets stay in the environment; the separate credential
// variables apply to any URI that embeds no userinfo of its own.
//
// Startup is fail-fast: every configured connection is verified with a ping
// (bounded by MONGO_PING_TIMEOUT) and an unreachable MongoDB aborts the
// application instead of deferring the failure to the first query.
package mongodb
