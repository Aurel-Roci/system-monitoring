build:
	go build -o /opt/system-monitor/bin/system-monitor cmd/main.go

run:
	go run cmd/main.go

test:
	go test ./...

clean:
	rm -rf bin/

sync-configs:
	sudo cp deployment/victoria-metrics/victoria-metrics.conf /etc/victoria-metrics/
	sudo cp deployment/system-monitoring/system-monitor.conf /etc/system-monitor/ 2>/dev/null || true
	sudo systemctl daemon-reload
	@echo "Configs synced. Restart services if needed."

# VictoriaMetrics management
setup:
	@echo "Setup VictoriaMetrics/System-Monitoring..."
	./deployment/install.sh

install-sys-monitor:
	./deployment/install-sys-monitor.sh

uninstall-vm:
	@echo "Uninstalling VictoriaMetrics..."
	./deployment/victoria-metrics/uninstall-victoriametrics.sh

restart-vm:
	sudo systemctl restart victoria-metrics
	@echo "VictoriaMetrics restarted"

logs-vm:
	sudo journalctl -u victoria-metrics -f

status-vm:
	sudo systemctl status victoria-metrics

status:
	@echo "=== System Monitor Status ==="
	@if sudo systemctl is-active --quiet system-monitor; then \
		echo "✅ System Monitor: Running"; \
	else \
		echo "❌ System Monitor: Stopped"; \
	fi
	@echo ""
	@echo "=== VictoriaMetrics Status ==="
	@if sudo systemctl is-active --quiet victoria-metrics; then \
		echo "✅ VictoriaMetrics: Running"; \
	else \
		echo "❌ VictoriaMetrics: Stopped"; \
	fi