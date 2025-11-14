import { Activity, Globe, Clock, AlertCircle } from "lucide-react";
import MetricCard from "@/components/MetricCard";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import StatusBadge from "@/components/StatusBadge";

const DNS = () => {
  const metrics = [
    { title: "Disponibilidade", value: "100", unit: "%", icon: Activity, status: "ok" as const },
    { title: "Tempo de Resposta", value: "12", unit: "ms", icon: Clock, status: "ok" as const, trend: "down" as const, trendValue: "-2ms" },
    { title: "Queries/segundo", value: "523", icon: Activity, status: "ok" as const, trend: "up" as const, trendValue: "+5%" },
    { title: "Taxa de Erro", value: "0", unit: "%", icon: AlertCircle, status: "ok" as const },
  ];

  const topDomains = [
    { domain: "api.example.com", queries: 12453, percentage: 45 },
    { domain: "www.example.com", queries: 8234, percentage: 30 },
    { domain: "cdn.example.com", queries: 4123, percentage: 15 },
    { domain: "mail.example.com", queries: 2745, percentage: 10 },
  ];

  return (
    <div className="space-y-6">
      <div className="flex items-start justify-between">
        <div className="flex items-center gap-3">
          <div className="rounded-lg bg-primary/10 p-3">
            <Globe className="h-8 w-8 text-primary" />
          </div>
          <div>
            <h2 className="text-3xl font-bold tracking-tight">DNS Server</h2>
            <p className="text-muted-foreground">Monitoramento de resolução de domínios</p>
          </div>
        </div>
        <StatusBadge status="ok" />
      </div>

      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
        {metrics.map((metric, index) => (
          <MetricCard key={index} {...metric} />
        ))}
      </div>

      <div className="grid gap-6 md:grid-cols-2">
        <Card>
          <CardHeader>
            <CardTitle>Domínios Mais Consultados</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="space-y-4">
              {topDomains.map((domain, index) => (
                <div key={index}>
                  <div className="flex items-center justify-between mb-2">
                    <code className="text-sm font-mono">{domain.domain}</code>
                    <span className="text-sm font-medium">{domain.queries.toLocaleString()}</span>
                  </div>
                  <div className="h-2 rounded-full bg-secondary overflow-hidden">
                    <div className="h-full bg-primary rounded-full transition-all" style={{ width: `${domain.percentage}%` }} />
                  </div>
                </div>
              ))}
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>Tipos de Query</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="space-y-3">
              <div className="flex items-center justify-between p-3 rounded-lg bg-secondary/50">
                <span className="text-sm font-medium">A Record</span>
                <span className="text-sm">18,234</span>
              </div>
              <div className="flex items-center justify-between p-3 rounded-lg bg-secondary/50">
                <span className="text-sm font-medium">AAAA Record</span>
                <span className="text-sm">5,432</span>
              </div>
              <div className="flex items-center justify-between p-3 rounded-lg bg-secondary/50">
                <span className="text-sm font-medium">MX Record</span>
                <span className="text-sm">1,234</span>
              </div>
              <div className="flex items-center justify-between p-3 rounded-lg bg-secondary/50">
                <span className="text-sm font-medium">TXT Record</span>
                <span className="text-sm">876</span>
              </div>
            </div>
          </CardContent>
        </Card>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>Performance de Resolução</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="grid gap-4 md:grid-cols-4">
            <div className="space-y-2">
              <p className="text-sm text-muted-foreground">Média</p>
              <p className="text-2xl font-bold">12<span className="text-base font-normal text-muted-foreground ml-1">ms</span></p>
            </div>
            <div className="space-y-2">
              <p className="text-sm text-muted-foreground">P50</p>
              <p className="text-2xl font-bold">8<span className="text-base font-normal text-muted-foreground ml-1">ms</span></p>
            </div>
            <div className="space-y-2">
              <p className="text-sm text-muted-foreground">P95</p>
              <p className="text-2xl font-bold">24<span className="text-base font-normal text-muted-foreground ml-1">ms</span></p>
            </div>
            <div className="space-y-2">
              <p className="text-sm text-muted-foreground">P99</p>
              <p className="text-2xl font-bold">45<span className="text-base font-normal text-muted-foreground ml-1">ms</span></p>
            </div>
          </div>
        </CardContent>
      </Card>
    </div>
  );
};

export default DNS;
