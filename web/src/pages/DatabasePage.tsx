import { Activity, Database as DatabaseIcon, HardDrive, Zap } from "lucide-react";
import MetricCard from "@/components/MetricCard";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import StatusBadge from "@/components/StatusBadge";

const DatabasePage = () => {
  const metrics = [
    { title: "Disponibilidade", value: "99.5", unit: "%", icon: Activity, status: "warning" as const },
    { title: "Queries/segundo", value: "856", icon: Zap, status: "ok" as const, trend: "up" as const, trendValue: "+8%" },
    { title: "Conexões Ativas", value: "42", icon: Activity, status: "ok" as const },
    { title: "Pool Utilizado", value: "85", unit: "%", icon: HardDrive, status: "warning" as const },
  ];

  const slowQueries = [
    { query: "SELECT * FROM users WHERE...", time: "2.3s", count: 12 },
    { query: "UPDATE orders SET status...", time: "1.8s", count: 8 },
    { query: "JOIN products ON orders...", time: "1.5s", count: 15 },
  ];

  return (
    <div className="space-y-6">
      <div className="flex items-start justify-between">
        <div className="flex items-center gap-3">
          <div className="rounded-lg bg-primary/10 p-3">
            <DatabaseIcon className="h-8 w-8 text-primary" />
          </div>
          <div>
            <h2 className="text-3xl font-bold tracking-tight">Database (SQL/NoSQL)</h2>
            <p className="text-muted-foreground">Monitoramento de banco de dados</p>
          </div>
        </div>
        <StatusBadge status="warning" />
      </div>

      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
        {metrics.map((metric, index) => (
          <MetricCard key={index} {...metric} />
        ))}
      </div>

      <div className="grid gap-6 md:grid-cols-2">
        <Card>
          <CardHeader>
            <CardTitle>Pool de Conexões</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="space-y-4">
              <div>
                <div className="flex justify-between mb-2">
                  <span className="text-sm text-muted-foreground">42 / 50 conexões</span>
                  <span className="text-sm font-medium text-status-warning">85%</span>
                </div>
                <div className="h-3 rounded-full bg-secondary overflow-hidden">
                  <div className="h-full bg-status-warning rounded-full transition-all" style={{ width: "85%" }} />
                </div>
              </div>
              <div className="grid grid-cols-2 gap-4 pt-2">
                <div>
                  <p className="text-xs text-muted-foreground">Ativas</p>
                  <p className="text-xl font-semibold">38</p>
                </div>
                <div>
                  <p className="text-xs text-muted-foreground">Idle</p>
                  <p className="text-xl font-semibold">4</p>
                </div>
              </div>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>Armazenamento</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="space-y-4">
              <div>
                <div className="flex justify-between mb-2">
                  <span className="text-sm text-muted-foreground">125 GB / 200 GB</span>
                  <span className="text-sm font-medium">62.5%</span>
                </div>
                <div className="h-3 rounded-full bg-secondary overflow-hidden">
                  <div className="h-full bg-primary rounded-full transition-all" style={{ width: "62.5%" }} />
                </div>
              </div>
              <div className="grid grid-cols-2 gap-4 pt-2">
                <div>
                  <p className="text-xs text-muted-foreground">Dados</p>
                  <p className="text-xl font-semibold">98 GB</p>
                </div>
                <div>
                  <p className="text-xs text-muted-foreground">Índices</p>
                  <p className="text-xl font-semibold">27 GB</p>
                </div>
              </div>
            </div>
          </CardContent>
        </Card>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>Queries Mais Lentas</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="space-y-3">
            {slowQueries.map((query, index) => (
              <div key={index} className="rounded-lg border border-border p-4 hover:bg-accent/50 transition-colors">
                <div className="flex items-start justify-between mb-2">
                  <code className="text-sm font-mono text-foreground">{query.query}</code>
                  <span className="text-sm font-medium text-status-warning">{query.time}</span>
                </div>
                <p className="text-xs text-muted-foreground">{query.count} execuções nas últimas 24h</p>
              </div>
            ))}
          </div>
        </CardContent>
      </Card>
    </div>
  );
};

export default DatabasePage;
