import { Server, Database, Globe, Mail, Activity, TrendingUp, AlertCircle } from "lucide-react";
import { Link } from "react-router-dom";
import ServiceCard from "@/components/ServiceCard";
import MetricCard from "@/components/MetricCard";
import ErrorState from "@/components/ErrorState";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import StatusBadge from "@/components/StatusBadge";
import { useHealth, useLatestMetrics } from "@/hooks/useMetrics";
import type { Status } from "@/lib/types";

const Dashboard = () => {
  const { data: health, isLoading: healthLoading, error: healthError, refetch: refetchHealth } = useHealth();
  const { data: latestMetrics, isLoading: metricsLoading, error: metricsError, refetch: refetchMetrics } = useLatestMetrics();

  const isLoading = healthLoading || metricsLoading;
  const hasError = healthError || metricsError;

  const getServiceStatus = (serviceName: string): Status => {
    const serviceMetrics = latestMetrics?.find((m) => m.service === serviceName);
    if (!serviceMetrics) return "critical";

    const upMetric = Object.keys(serviceMetrics.metrics).find((key) =>
      key.includes("_up")
    );

    if (upMetric && serviceMetrics.metrics[upMetric] === 0) return "critical";

    const latencyMetric = Object.keys(serviceMetrics.metrics).find((key) =>
      key.includes("latency") || key.includes("duration")
    );

    if (latencyMetric && serviceMetrics.metrics[latencyMetric] > 1000) return "warning";

    return "ok";
  };

  const getServiceUptime = (serviceName: string): string => {
    const serviceMetrics = latestMetrics?.find((m) => m.service === serviceName);
    if (!serviceMetrics) return "0%";

    const upMetric = Object.keys(serviceMetrics.metrics).find((key) =>
      key.includes("_up")
    );

    if (upMetric) {
      return serviceMetrics.metrics[upMetric] === 1 ? "100%" : "0%";
    }

    return "N/A";
  };

  const getServiceMetrics = (serviceName: string) => {
    const serviceMetrics = latestMetrics?.find((m) => m.service === serviceName);
    if (!serviceMetrics) return [];

    const metrics = [];

    if (serviceName === "http" || serviceName === "https") {
      const latency = serviceMetrics.metrics.http_latency_ms;
      const statusCode = serviceMetrics.metrics.http_status_code;

      if (latency !== undefined) {
        metrics.push({ label: "Latência", value: `${latency.toFixed(0)}ms` });
      }
      if (statusCode !== undefined) {
        metrics.push({ label: "Status Code", value: statusCode.toString() });
      }
    } else if (serviceName === "postgres") {
      const connections = serviceMetrics.metrics.postgres_active_connections;
      const latency = serviceMetrics.metrics.postgres_query_latency_ms;

      if (connections !== undefined) {
        metrics.push({ label: "Conexões Ativas", value: connections.toString() });
      }
      if (latency !== undefined) {
        metrics.push({ label: "Latência Query", value: `${latency.toFixed(0)}ms` });
      }
    } else if (serviceName === "dns") {
      const latency = serviceMetrics.metrics.dns_lookup_duration_ms;

      if (latency !== undefined) {
        metrics.push({ label: "Tempo de Lookup", value: `${latency.toFixed(0)}ms` });
      }
    } else if (serviceName === "smtp") {
      const latency = serviceMetrics.metrics.smtp_handshake_duration_ms;

      if (latency !== undefined) {
        metrics.push({ label: "Tempo Handshake", value: `${latency.toFixed(0)}ms` });
      }
    }

    while (metrics.length < 3) {
      metrics.push({ label: "-", value: "-" });
    }

    return metrics.slice(0, 3);
  };

  const services = [
    {
      title: "Web Server",
      icon: Server,
      status: getServiceStatus("http"),
      uptime: getServiceUptime("http"),
      metrics: getServiceMetrics("http"),
      link: "/webserver",
    },
    {
      title: "Database",
      icon: Database,
      status: getServiceStatus("postgres"),
      uptime: getServiceUptime("postgres"),
      metrics: getServiceMetrics("postgres"),
      link: "/database",
    },
    {
      title: "DNS",
      icon: Globe,
      status: getServiceStatus("dns"),
      uptime: getServiceUptime("dns"),
      metrics: getServiceMetrics("dns"),
      link: "/dns",
    },
    {
      title: "SMTP",
      icon: Mail,
      status: getServiceStatus("smtp"),
      uptime: getServiceUptime("smtp"),
      metrics: getServiceMetrics("smtp"),
      link: "/smtp",
    },
  ];

  const activeAlerts = health?.active_alerts || [];
  const activeServicesCount = services.filter((s) => s.status !== "critical").length;
  const totalServices = services.length;

  if (hasError && !isLoading) {
    return (
      <div className="space-y-6">
        <div>
          <h2 className="text-3xl font-bold tracking-tight mb-1">Dashboard de Monitoramento</h2>
          <p className="text-muted-foreground">Visão consolidada de todos os serviços</p>
        </div>
        <ErrorState
          onRetry={() => {
            refetchHealth();
            refetchMetrics();
          }}
        />
      </div>
    );
  }

  if (isLoading) {
    return (
      <div className="flex items-center justify-center min-h-[400px]">
        <div className="text-center">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-primary mx-auto" />
          <p className="mt-4 text-muted-foreground">Carregando dashboard...</p>
        </div>
      </div>
    );
  }

  const avgLatency =
    latestMetrics
      ?.map((m) => {
        const latencyKey = Object.keys(m.metrics).find(
          (k) => k.includes("latency") || k.includes("duration")
        );
        return latencyKey ? m.metrics[latencyKey] : 0;
      })
      .reduce((sum, val) => sum + val, 0) /
      (latestMetrics?.length || 1) || 0;

  return (
    <div className="space-y-6">
      <div>
        <h2 className="text-3xl font-bold tracking-tight mb-1">Dashboard de Monitoramento</h2>
        <p className="text-muted-foreground">Visão consolidada de todos os serviços</p>
      </div>

      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
        <MetricCard
          title="Serviços Ativos"
          value={`${activeServicesCount}/${totalServices}`}
          icon={Activity}
          status={activeServicesCount === totalServices ? "ok" : "warning"}
          trend="stable"
        />
        <MetricCard
          title="Total de Métricas"
          value={health?.metrics_count || 0}
          icon={TrendingUp}
          status="ok"
        />
        <MetricCard
          title="Alertas Ativos"
          value={activeAlerts.length}
          icon={AlertCircle}
          status={activeAlerts.length === 0 ? "ok" : "warning"}
        />
        <MetricCard
          title="Latência Média"
          value={avgLatency.toFixed(0)}
          unit="ms"
          icon={Activity}
          status={avgLatency < 100 ? "ok" : avgLatency < 500 ? "warning" : "critical"}
        />
      </div>

      {activeAlerts.length > 0 && (
        <Card>
          <CardHeader>
            <div className="flex items-center justify-between">
              <CardTitle>Alertas Ativos</CardTitle>
              <Link
                to="/alerts"
                className="text-sm text-primary hover:underline"
              >
                Ver todos
              </Link>
            </div>
          </CardHeader>
          <CardContent>
            <div className="space-y-3">
              {activeAlerts.slice(0, 5).map((alert, index) => (
                <div
                  key={index}
                  className="flex items-start justify-between rounded-lg border border-border p-4 transition-colors hover:bg-accent/50"
                >
                  <div className="flex-1">
                    <div className="flex items-center gap-2 mb-1">
                      <StatusBadge status={alert.severity as Status} />
                      <span className="font-medium">{alert.service}</span>
                      {alert.target && (
                        <span className="text-sm text-muted-foreground">({alert.target})</span>
                      )}
                    </div>
                    <p className="text-sm text-muted-foreground">{alert.message}</p>
                  </div>
                  <span className="text-xs text-muted-foreground whitespace-nowrap">
                    {new Date(alert.triggered_at).toLocaleTimeString("pt-BR")}
                  </span>
                </div>
              ))}
            </div>
          </CardContent>
        </Card>
      )}

      <div>
        <h3 className="text-xl font-semibold mb-4">Serviços Monitorados</h3>
        <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-2">
          {services.map((service, index) => (
            <ServiceCard key={index} {...service} />
          ))}
        </div>
      </div>

      {health && (
        <Card>
          <CardHeader>
            <CardTitle>Informações do Sistema</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
              <div>
                <p className="text-sm text-muted-foreground">Status</p>
                <p className="text-lg font-semibold capitalize">{health.status}</p>
              </div>
              <div>
                <p className="text-sm text-muted-foreground">Uptime</p>
                <p className="text-lg font-semibold">
                  {Math.floor(health.uptime / 3600)}h {Math.floor((health.uptime % 3600) / 60)}m
                </p>
              </div>
              <div>
                <p className="text-sm text-muted-foreground">Targets</p>
                <p className="text-lg font-semibold">{health.targets_count}</p>
              </div>
              <div>
                <p className="text-sm text-muted-foreground">Última Ingestão</p>
                <p className="text-lg font-semibold">
                  {new Date(health.last_ingest).toLocaleTimeString("pt-BR")}
                </p>
              </div>
            </div>
          </CardContent>
        </Card>
      )}
    </div>
  );
};

export default Dashboard;
