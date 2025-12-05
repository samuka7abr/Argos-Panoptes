import { Shield, AlertTriangle, TrendingUp, Lock } from "lucide-react";
import MetricCard from "@/components/MetricCard";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import StatusBadge from "@/components/StatusBadge";
import { useQuery } from "@tanstack/react-query";
import { api } from "@/lib/api";
import { formatDistanceToNow } from "date-fns";
import { ptBR } from "date-fns/locale";

const Security = () => {
  const { data: stats, isLoading: statsLoading } = useQuery({
    queryKey: ["security", "stats"],
    queryFn: () => api.security.stats(),
    refetchInterval: 30000,
  });

  const { data: events, isLoading: eventsLoading } = useQuery({
    queryKey: ["security", "events"],
    queryFn: () => api.security.events(10),
    refetchInterval: 30000,
  });

  const { data: failedLogins, isLoading: loginsLoading } = useQuery({
    queryKey: ["security", "failed-logins"],
    queryFn: () => api.security.failedLogins(10),
    refetchInterval: 30000,
  });

  const { data: vulnerabilities, isLoading: vulnsLoading } = useQuery({
    queryKey: ["security", "vulnerabilities"],
    queryFn: () => api.security.vulnerabilities(),
    refetchInterval: 60000,
  });

  const isLoading = statsLoading || eventsLoading || loginsLoading || vulnsLoading;

  const metrics = [
    {
      title: "Tentativas de Login Falhadas",
      value: stats?.failed_logins?.toString() || "0",
      icon: Lock,
      status: (stats?.failed_logins || 0) > 10 ? ("warning" as const) : ("ok" as const),
      trend: (stats?.failed_logins || 0) > 0 ? ("up" as const) : ("stable" as const),
      trendValue: stats?.failed_logins ? `+${stats.failed_logins}` : "0",
    },
    {
      title: "Anomalias de Tráfego",
      value: stats?.traffic_anomalies?.toString() || "0",
      icon: TrendingUp,
      status: (stats?.traffic_anomalies || 0) > 0 ? ("warning" as const) : ("ok" as const),
    },
    {
      title: "Alterações de Config",
      value: stats?.config_changes?.toString() || "0",
      icon: AlertTriangle,
      status: (stats?.config_changes || 0) > 0 ? ("warning" as const) : ("ok" as const),
    },
    {
      title: "Vulnerabilidades",
      value: stats?.vulnerabilities?.toString() || "0",
      icon: Shield,
      status:
        (stats?.vulnerabilities || 0) > 0
          ? (vulnerabilities?.vulnerabilities?.some((v) => v.severity === "high" || v.severity === "critical")
              ? ("critical" as const)
              : ("warning" as const))
          : ("ok" as const),
    },
  ];

  if (isLoading) {
    return (
      <div className="flex items-center justify-center min-h-[400px]">
        <div className="text-center">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-primary mx-auto" />
          <p className="mt-4 text-muted-foreground">Carregando dados de segurança...</p>
        </div>
      </div>
    );
  }

  const getSeverityStatus = (severity: string): "ok" | "warning" | "critical" => {
    if (severity === "critical" || severity === "high") return "critical";
    if (severity === "warning" || severity === "medium") return "warning";
    return "ok";
  };

  return (
    <div className="space-y-6">
      <div className="flex items-start justify-between">
        <div className="flex items-center gap-3">
          <div className="rounded-lg bg-primary/10 p-3">
            <Shield className="h-8 w-8 text-primary" />
          </div>
          <div>
            <h2 className="text-3xl font-bold tracking-tight">Segurança e Monitoramento</h2>
            <p className="text-muted-foreground">Detecção de anomalias e vulnerabilidades</p>
          </div>
        </div>
        <StatusBadge
          status={
            (stats?.failed_logins || 0) > 10 ||
            (stats?.traffic_anomalies || 0) > 0 ||
            (stats?.vulnerabilities || 0) > 0
              ? "warning"
              : "ok"
          }
        />
      </div>

      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
        {metrics.map((metric, index) => (
          <MetricCard key={index} {...metric} />
        ))}
      </div>

      <Card>
        <CardHeader>
          <CardTitle>Eventos de Segurança Recentes</CardTitle>
        </CardHeader>
        <CardContent>
          {eventsLoading ? (
            <div className="flex items-center justify-center py-8">
              <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary" />
            </div>
          ) : events?.events && events.events.length > 0 ? (
            <div className="space-y-3">
              {events.events.map((event) => (
                <div
                  key={event.id}
                  className="rounded-lg border border-border p-4 hover:bg-accent/50 transition-colors"
                >
                  <div className="flex items-start justify-between mb-2">
                    <div className="flex items-center gap-2">
                      <StatusBadge status={getSeverityStatus(event.severity)} />
                      <span className="font-medium">{event.type}</span>
                    </div>
                    <span className="text-xs text-muted-foreground">
                      {formatDistanceToNow(new Date(event.created_at), {
                        addSuffix: true,
                        locale: ptBR,
                      })}
                    </span>
                  </div>
                  <p className="text-sm text-muted-foreground mb-1">{event.description}</p>
                  {event.service && (
                    <p className="text-xs text-muted-foreground">Serviço: {event.service}</p>
                  )}
                </div>
              ))}
            </div>
          ) : (
            <div className="flex flex-col items-center justify-center py-12 text-center">
              <Shield className="h-12 w-12 text-muted-foreground mb-4" />
              <p className="text-lg font-medium">Nenhum evento de segurança registrado</p>
              <p className="text-sm text-muted-foreground mt-2">
                Eventos de segurança aparecerão aqui quando detectados
              </p>
            </div>
          )}
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>Vulnerabilidades Conhecidas</CardTitle>
        </CardHeader>
        <CardContent>
          {vulnsLoading ? (
            <div className="flex items-center justify-center py-8">
              <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary" />
            </div>
          ) : vulnerabilities?.vulnerabilities && vulnerabilities.vulnerabilities.length > 0 ? (
            <div className="space-y-3">
              {vulnerabilities.vulnerabilities.map((vuln) => (
                <div key={vuln.id} className="rounded-lg border border-border p-4">
                  <div className="flex items-start justify-between mb-2">
                    <div>
                      <div className="flex items-center gap-2 mb-1">
                        {vuln.cve && (
                          <code className="text-xs font-mono bg-secondary px-2 py-1 rounded">
                            {vuln.cve}
                          </code>
                        )}
                        <span
                          className={`text-xs px-2 py-1 rounded font-medium ${
                            vuln.severity === "high" || vuln.severity === "critical"
                              ? "bg-status-critical/20 text-status-critical"
                              : "bg-status-warning/20 text-status-warning"
                          }`}
                        >
                          {vuln.severity.toUpperCase()}
                        </span>
                      </div>
                      <p className="text-sm font-medium">{vuln.service}</p>
                    </div>
                  </div>
                  {vuln.description && (
                    <p className="text-sm text-muted-foreground">{vuln.description}</p>
                  )}
                </div>
              ))}
            </div>
          ) : (
            <div className="flex flex-col items-center justify-center py-12 text-center">
              <Shield className="h-12 w-12 text-muted-foreground mb-4" />
              <p className="text-lg font-medium">Nenhuma vulnerabilidade detectada</p>
              <p className="text-sm text-muted-foreground mt-2">
                Vulnerabilidades conhecidas aparecerão aqui
              </p>
            </div>
          )}
        </CardContent>
      </Card>

      <div className="grid gap-6 md:grid-cols-2">
        <Card>
          <CardHeader>
            <CardTitle>Top IPs com Falhas de Autenticação</CardTitle>
          </CardHeader>
          <CardContent>
            {loginsLoading ? (
              <div className="flex items-center justify-center py-8">
                <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary" />
              </div>
            ) : failedLogins?.by_ip && failedLogins.by_ip.length > 0 ? (
              <div className="space-y-3">
                {failedLogins.by_ip.map((item, index) => (
                  <div
                    key={index}
                    className="flex items-center justify-between p-3 rounded-lg bg-secondary/50"
                  >
                    <code className="text-sm font-mono">{item.ip_address}</code>
                    <span className="text-sm font-medium text-status-warning">
                      {item.count} tentativa{item.count > 1 ? "s" : ""}
                    </span>
                  </div>
                ))}
              </div>
            ) : (
              <div className="flex flex-col items-center justify-center py-8 text-center">
                <Lock className="h-8 w-8 text-muted-foreground mb-2" />
                <p className="text-sm text-muted-foreground">Nenhuma tentativa falhada registrada</p>
              </div>
            )}
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>Análise de Tráfego</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="space-y-4">
              <div>
                <div className="flex justify-between mb-2">
                  <span className="text-sm text-muted-foreground">Requests normais</span>
                  <span className="text-sm font-medium">
                    {stats
                      ? `${Math.max(0, 100 - ((stats.traffic_anomalies || 0) * 10)).toFixed(0)}%`
                      : "100%"}
                  </span>
                </div>
                <div className="h-2 rounded-full bg-secondary overflow-hidden">
                  <div
                    className="h-full bg-status-ok rounded-full"
                    style={{
                      width: `${
                        stats
                          ? Math.max(0, 100 - ((stats.traffic_anomalies || 0) * 10))
                          : 100
                      }%`,
                    }}
                  />
                </div>
              </div>
              <div>
                <div className="flex justify-between mb-2">
                  <span className="text-sm text-muted-foreground">Suspeitos</span>
                  <span className="text-sm font-medium text-status-warning">
                    {stats ? `${Math.min(100, (stats.traffic_anomalies || 0) * 10).toFixed(0)}%` : "0%"}
                  </span>
                </div>
                <div className="h-2 rounded-full bg-secondary overflow-hidden">
                  <div
                    className="h-full bg-status-warning rounded-full"
                    style={{
                      width: `${
                        stats ? Math.min(100, (stats.traffic_anomalies || 0) * 10) : 0
                      }%`,
                    }}
                  />
                </div>
              </div>
            </div>
          </CardContent>
        </Card>
      </div>
    </div>
  );
};

export default Security;
