import { Shield, AlertTriangle, TrendingUp, Lock } from "lucide-react";
import MetricCard from "@/components/MetricCard";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import StatusBadge from "@/components/StatusBadge";

const Security = () => {
  const metrics = [
    { title: "Tentativas de Login Falhadas", value: "12", icon: Lock, status: "warning" as const, trend: "up" as const, trendValue: "+4" },
    { title: "Anomalias de Tráfego", value: "3", icon: TrendingUp, status: "warning" as const },
    { title: "Alterações de Config", value: "0", icon: AlertTriangle, status: "ok" as const },
    { title: "Vulnerabilidades", value: "2", icon: Shield, status: "warning" as const },
  ];

  const securityEvents = [
    {
      type: "Brute Force Detectado",
      severity: "warning" as const,
      description: "5 tentativas de login falhadas do IP 192.168.1.100",
      time: "8 min atrás",
      service: "Web Server",
    },
    {
      type: "Pico de Tráfego Anormal",
      severity: "warning" as const,
      description: "Aumento súbito de 300% nas requisições",
      time: "15 min atrás",
      service: "Web Server",
    },
    {
      type: "Possível DDoS",
      severity: "critical" as const,
      description: "10,000+ requisições de múltiplos IPs em 30 segundos",
      time: "45 min atrás",
      service: "DNS",
    },
  ];

  const vulnerabilities = [
    {
      service: "Web Server",
      cve: "CVE-2024-1234",
      severity: "medium",
      description: "Vulnerabilidade no OpenSSL 1.1.1k",
    },
    {
      service: "Database",
      cve: "CVE-2024-5678",
      severity: "high",
      description: "SQL Injection potencial na versão PostgreSQL 12.1",
    },
  ];

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
        <StatusBadge status="warning" />
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
          <div className="space-y-3">
            {securityEvents.map((event, index) => (
              <div key={index} className="rounded-lg border border-border p-4 hover:bg-accent/50 transition-colors">
                <div className="flex items-start justify-between mb-2">
                  <div className="flex items-center gap-2">
                    <StatusBadge status={event.severity} />
                    <span className="font-medium">{event.type}</span>
                  </div>
                  <span className="text-xs text-muted-foreground">{event.time}</span>
                </div>
                <p className="text-sm text-muted-foreground mb-1">{event.description}</p>
                <p className="text-xs text-muted-foreground">Serviço: {event.service}</p>
              </div>
            ))}
          </div>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>Vulnerabilidades Conhecidas</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="space-y-3">
            {vulnerabilities.map((vuln, index) => (
              <div key={index} className="rounded-lg border border-border p-4">
                <div className="flex items-start justify-between mb-2">
                  <div>
                    <div className="flex items-center gap-2 mb-1">
                      <code className="text-xs font-mono bg-secondary px-2 py-1 rounded">{vuln.cve}</code>
                      <span className={`text-xs px-2 py-1 rounded font-medium ${
                        vuln.severity === "high" ? "bg-status-critical/20 text-status-critical" : "bg-status-warning/20 text-status-warning"
                      }`}>
                        {vuln.severity.toUpperCase()}
                      </span>
                    </div>
                    <p className="text-sm font-medium">{vuln.service}</p>
                  </div>
                </div>
                <p className="text-sm text-muted-foreground">{vuln.description}</p>
              </div>
            ))}
          </div>
        </CardContent>
      </Card>

      <div className="grid gap-6 md:grid-cols-2">
        <Card>
          <CardHeader>
            <CardTitle>Top IPs com Falhas de Autenticação</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="space-y-3">
              <div className="flex items-center justify-between p-3 rounded-lg bg-secondary/50">
                <code className="text-sm font-mono">192.168.1.100</code>
                <span className="text-sm font-medium text-status-warning">5 tentativas</span>
              </div>
              <div className="flex items-center justify-between p-3 rounded-lg bg-secondary/50">
                <code className="text-sm font-mono">10.0.0.42</code>
                <span className="text-sm font-medium text-status-warning">4 tentativas</span>
              </div>
              <div className="flex items-center justify-between p-3 rounded-lg bg-secondary/50">
                <code className="text-sm font-mono">172.16.0.88</code>
                <span className="text-sm font-medium text-status-warning">3 tentativas</span>
              </div>
            </div>
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
                  <span className="text-sm font-medium">92%</span>
                </div>
                <div className="h-2 rounded-full bg-secondary overflow-hidden">
                  <div className="h-full bg-status-ok rounded-full" style={{ width: "92%" }} />
                </div>
              </div>
              <div>
                <div className="flex justify-between mb-2">
                  <span className="text-sm text-muted-foreground">Suspeitos</span>
                  <span className="text-sm font-medium text-status-warning">8%</span>
                </div>
                <div className="h-2 rounded-full bg-secondary overflow-hidden">
                  <div className="h-full bg-status-warning rounded-full" style={{ width: "8%" }} />
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
