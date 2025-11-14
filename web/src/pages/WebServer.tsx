import { Activity, Clock, TrendingUp, AlertCircle, Server } from "lucide-react";
import MetricCard from "@/components/MetricCard";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import StatusBadge from "@/components/StatusBadge";

const WebServer = () => {
  // Mock data - será integrado com backend Golang
  const metrics = [
    { title: "Disponibilidade", value: "99.9", unit: "%", icon: Activity, status: "ok" as const, trend: "stable" as const },
    { title: "Requests/segundo", value: "1,234", icon: TrendingUp, status: "ok" as const, trend: "up" as const, trendValue: "+12%" },
    { title: "Latência Média", value: "45", unit: "ms", icon: Clock, status: "ok" as const, trend: "down" as const, trendValue: "-5ms" },
    { title: "Taxa de Erro", value: "0.01", unit: "%", icon: AlertCircle, status: "ok" as const },
  ];

  const errorCodes = [
    { code: "200", count: 125847, percentage: 99.85 },
    { code: "404", count: 156, percentage: 0.12 },
    { code: "500", count: 12, percentage: 0.01 },
    { code: "502", count: 8, percentage: 0.01 },
    { code: "503", count: 3, percentage: 0.00 },
  ];

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-start justify-between">
        <div className="flex items-center gap-3">
          <div className="rounded-lg bg-primary/10 p-3">
            <Server className="h-8 w-8 text-primary" />
          </div>
          <div>
            <h2 className="text-3xl font-bold tracking-tight">Web Server (HTTP/HTTPS)</h2>
            <p className="text-muted-foreground">Monitoramento de servidor web e APIs</p>
          </div>
        </div>
        <StatusBadge status="ok" />
      </div>

      {/* Metrics Grid */}
      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
        {metrics.map((metric, index) => (
          <MetricCard key={index} {...metric} />
        ))}
      </div>

      {/* Additional Stats */}
      <div className="grid gap-6 md:grid-cols-2">
        <Card>
          <CardHeader>
            <CardTitle>Códigos de Status HTTP</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="space-y-3">
              {errorCodes.map((error) => (
                <div key={error.code} className="flex items-center justify-between">
                  <div className="flex items-center gap-3">
                    <span className={`font-mono text-sm px-2 py-1 rounded ${
                      error.code.startsWith("2") ? "bg-status-ok/20 text-status-ok" :
                      error.code.startsWith("4") ? "bg-status-warning/20 text-status-warning" :
                      "bg-status-critical/20 text-status-critical"
                    }`}>
                      {error.code}
                    </span>
                    <span className="text-sm text-muted-foreground">{error.count.toLocaleString()} requisições</span>
                  </div>
                  <span className="text-sm font-medium">{error.percentage.toFixed(2)}%</span>
                </div>
              ))}
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>Conexões Ativas</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="space-y-4">
              <div>
                <div className="flex items-center justify-between mb-2">
                  <span className="text-sm text-muted-foreground">Total</span>
                  <span className="text-2xl font-bold">342</span>
                </div>
                <div className="h-2 rounded-full bg-secondary overflow-hidden">
                  <div className="h-full bg-primary rounded-full" style={{ width: "68%" }} />
                </div>
              </div>
              <div className="grid grid-cols-2 gap-4 pt-2">
                <div>
                  <p className="text-xs text-muted-foreground">Keep-Alive</p>
                  <p className="text-xl font-semibold">287</p>
                </div>
                <div>
                  <p className="text-xs text-muted-foreground">Novas</p>
                  <p className="text-xl font-semibold">55</p>
                </div>
              </div>
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Performance Details */}
      <Card>
        <CardHeader>
          <CardTitle>Detalhes de Performance</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="grid gap-4 md:grid-cols-3">
            <div className="space-y-2">
              <p className="text-sm text-muted-foreground">Latência P50</p>
              <p className="text-2xl font-bold">32<span className="text-base font-normal text-muted-foreground ml-1">ms</span></p>
            </div>
            <div className="space-y-2">
              <p className="text-sm text-muted-foreground">Latência P95</p>
              <p className="text-2xl font-bold">78<span className="text-base font-normal text-muted-foreground ml-1">ms</span></p>
            </div>
            <div className="space-y-2">
              <p className="text-sm text-muted-foreground">Latência P99</p>
              <p className="text-2xl font-bold">145<span className="text-base font-normal text-muted-foreground ml-1">ms</span></p>
            </div>
          </div>
        </CardContent>
      </Card>
    </div>
  );
};

export default WebServer;
