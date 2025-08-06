#!/bin/bash
# Demo script for Peep

set -e

echo "🎪 Peep Demo - Observability for humans"
echo "======================================="
echo

# Clean start
echo "🧹 Cleaning previous data..."
rm -f logs.db
echo

# Build the latest version
echo "🔨 Building Peep..."
make build
echo

# Demo 1: JSON logs via stdin
echo "📥 Demo 1: Ingesting JSON logs via stdin"
echo '{"timestamp":"2023-08-06T10:30:45Z","level":"info","message":"Application started","service":"api"}' | ./peep
echo '{"timestamp":"2023-08-06T10:30:46Z","level":"error","message":"Database connection failed","service":"api"}' | ./peep
echo '{"timestamp":"2023-08-06T10:30:47Z","level":"debug","message":"Processing user request","service":"api","user_id":1234}' | ./peep
echo

# Demo 2: Common format logs
echo "📥 Demo 2: Ingesting common format logs"
echo '2023-08-06 10:30:48 WARN [cache] Memory usage above 80%' | ./peep
echo '2023-08-06 10:30:49 INFO [web] New user registered' | ./peep
echo

# Demo 3: File ingestion
echo "📥 Demo 3: Ingesting from file"
./peep ingest sample.log
echo

# Demo 4: List logs
echo "📋 Demo 4: Viewing stored logs"
./peep list
echo

# Demo 5: List with custom limit
echo "📋 Demo 5: Viewing last 5 logs"
./peep list --limit 5
echo

echo "✅ Demo complete!"
echo "💡 Try these commands:"
echo "   ./peep --help              # Show all commands"
echo "   ./peep list                # View recent logs"
echo "   ./peep tui                 # Start TUI (coming soon)"
echo "   ./peep web                 # Start web interface (coming soon)"
echo "   docker logs myapp | ./peep # Real-time log ingestion"
