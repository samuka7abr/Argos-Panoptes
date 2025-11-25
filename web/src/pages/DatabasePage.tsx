import { Database, Activity, Clock, Users } from "lucide-react";
import MetricCard from "@/components/MetricCard";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import StatusBadge from "@/components/StatusBadge";
import { useLatestMetrics, useMetricQuery } from "@/hooks/useMetrics";
import {
  LineChart,
  Line,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
  Legend,
} from "recharts";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { useState } from "react";

const DatabasePage = () => {
  const { data: latestMetrics, isLoading } = useLatestMetrics();
  const postgresMetrics = latestMetrics?.filter((m) => m.service === "postgres");
  const [selectedTarget, setSelectedTarget] = useState<string>("");

  const currentTarget = selectedTarget || postgresMetrics?.[0]?.target || "";

  const { data: latencyData } = useMetricQuery(
    "postgres",
    currentTarget,
    "postgres_query_latency_ms",
    "1h",
    !!currentTarget
  );

  const { data: connectionsData } = useMetricQuery(
    "postgres",
    currentTarget,
    "postgres_active_connections",
    "1h",
    !!currentTarget
  );

  const currentMetrics = postgresMetrics?.find((m) => m.target === currentTarget);

  const isUp = currentMetrics?.metrics.postgres_up === 1;
  const latency = currentMetrics?.metrics.postgres_query_latency_ms || 0;
  const connections = currentMetrics?.metrics.postgres_active_connections || 0;
  const maxConnections = currentMetrics?.metrics.postgres_max_connections || 100;

  const latencyChartData =
    latencyData?.data.map((d) => ({
      time: new Date(d.timestamp).toLocaleTimeString("pt-BR", {
        hour: "2-digit",
        minute: "2-digit",
      }),
      latency: d.value,
    })) || [];

  const connectionsChartData =
    connectionsData?.data.map((d) => ({
      time: new Date(d.timestamp).toLocaleTimeString("pt-BR", {
        hour: "2-digit",
        minute: "2-digit",
      }),
      connections: d.value,
    })) || [];

  if (isLoading) {
    return (
      <div className="flex items-center justify-center min-h-[400px]">
        <div className="text-center">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-primary mx-auto" />
          <p className="mt-4 text-muted-foreground">Carregando métricas...</p>
        </div>
      </div>
    );
  }

  if (!postgresMetrics || postgresMetrics.length === 0) {
    return (
      <div className="space-y-6">
        <div className="flex items-center gap-3">
          <div className="rounded-lg bg-primary/10 p-3">
            <Database className="h-8 w-8 text-primary" />
          </div>
          <div>
            <h2 className="text-3xl font-bold tracking-tight">Database (PostgreSQL)</h2>
            <p className="text-muted-foreground">Monitoramento de banco de dados</p>
          </div>
        </div>
        <Card>
          <CardContent className="flex flex-col items-center justify-center py-12">
            <Database className="h-12 w-12 text-muted-foreground mb-4" />
            <p className="text-lg font-medium">Nenhum banco de dados monitorado</p>
            <p className="text-sm text-muted-foreground mt-2">
              Configure o Agent para monitorar bancos PostgreSQL
            </p>
          </CardContent>
        </Card>
      </div>
    );
  }

  const connectionUsage = maxConnections > 0 ? (connections / maxConnections) * 100 : 0;

  return (
    <div className="space-y-6">
      <div className="flex items-start justify-between">
        <div className="flex items-center gap-3">
          <div className="rounded-lg bg-primary/10 p-3">
            <Database className="h-8 w-8 text-primary" />
          </div>
          <div>
            <h2 className="text-3xl font-bold tracking-tight">Database (PostgreSQL)</h2>
            <p className="text-muted-foreground">Monitoramento de banco de dados</p>
          </div>
        </div>
        <StatusBadge status={isUp ? "ok" : "critical"} />
      </div>

      {postgresMetrics.length > 1 && (
        <div className="flex items-center gap-3">
          <label className="text-sm font-medium">Target:</label>
          <Select value={currentTarget} onValueChange={setSelectedTarget}>
            <SelectTrigger className="w-[300px]">
              <SelectValue placeholder="Selecione um target" />
            </SelectTrigger>
            <SelectContent>
              {postgresMetrics.map((m) => (
                <SelectItem key={m.target} value={m.target}>
                  {m.target}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>
      )}

      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
        <MetricCard
          title="Disponibilidade"
          value={isUp ? "100" : "0"}
          unit="%"
          icon={Activity}
          status={isUp ? "ok" : "critical"}
        />
        <MetricCard
          title="Latência de Query"
          value={latency.toFixed(0)}
          unit="ms"
          icon={Clock}
          status={latency < 50 ? "ok" : latency < 200 ? "warning" : "critical"}
        />
        <MetricCard
          title="Conexões Ativas"
          value={connections}
          icon={Users}
          status={connectionUsage < 70 ? "ok" : connectionUsage < 90 ? "warning" : "critical"}
        />
        <MetricCard
          title="Uso do Pool"
          value={connectionUsage.toFixed(0)}
          unit="%"
          icon={Activity}
          status={connectionUsage < 70 ? "ok" : connectionUsage < 90 ? "warning" : "critical"}
        />
      </div>

      <div className="grid gap-6 md:grid-cols-2">
        <Card>
          <CardHeader>
            <CardTitle>Latência de Queries (última hora)</CardTitle>
          </CardHeader>
          <CardContent>
            {latencyChartData.length > 0 ? (
              <ResponsiveContainer width="100%" height={250}>
                <LineChart data={latencyChartData}>
                  <CartesianGrid strokeDasharray="3 3" />
                  <XAxis dataKey="time" />
                  <YAxis />
                  <Tooltip />
                  <Legend />
                  <Line
                    type="monotone"
                    dataKey="latency"
                    stroke="hsl(var(--primary))"
                    name="Latência (ms)"
                    strokeWidth={2}
                  />
                </LineChart>
              </ResponsiveContainer>
            ) : (
              <div className="flex items-center justify-center h-[250px] text-muted-foreground">
                Sem dados disponíveis
              </div>
            )}
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>Conexões Ativas (última hora)</CardTitle>
          </CardHeader>
          <CardContent>
            {connectionsChartData.length > 0 ? (
              <ResponsiveContainer width="100%" height={250}>
                <LineChart data={connectionsChartData}>
                  <CartesianGrid strokeDasharray="3 3" />
                  <XAxis dataKey="time" />
                  <YAxis />
                  <Tooltip />
                  <Legend />
                  <Line
                    type="monotone"
                    dataKey="connections"
                    stroke="hsl(var(--chart-2))"
                    name="Conexões"
                    strokeWidth={2}
                  />
                </LineChart>
              </ResponsiveContainer>
            ) : (
              <div className="flex items-center justify-center h-[250px] text-muted-foreground">
                Sem dados disponíveis
              </div>
            )}
          </CardContent>
        </Card>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>Pool de Conexões</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="space-y-4">
            <div>
              <div className="flex items-center justify-between mb-2">
                <span className="text-sm text-muted-foreground">Uso Atual</span>
                <span className="text-2xl font-bold">
                  {connections} / {maxConnections}
                </span>
              </div>
              <div className="h-3 rounded-full bg-secondary overflow-hidden">
                <div
                  className={`h-full rounded-full transition-all ${
                    connectionUsage < 70
                      ? "bg-status-ok"
                      : connectionUsage < 90
                        ? "bg-status-warning"
                        : "bg-status-critical"
                  }`}
                  style={{ width: `${connectionUsage}%` }}
                />
              </div>
            </div>
          </div>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>Todas as Instâncias</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="space-y-3">
            {postgresMetrics.map((metric) => {
              const targetUp = metric.metrics.postgres_up === 1;
              const targetLatency = metric.metrics.postgres_query_latency_ms || 0;
              const targetConnections = metric.metrics.postgres_active_connections || 0;

              return (
                <div
                  key={metric.target}
                  className="flex items-center justify-between rounded-lg border p-3"
                >
                  <div className="flex-1">
                    <div className="flex items-center gap-2 mb-1">
                      <StatusBadge status={targetUp ? "ok" : "critical"} />
                      <span className="font-medium">{metric.target}</span>
                    </div>
                    <div className="flex gap-4 text-xs text-muted-foreground">
                      <span>Latência: {targetLatency.toFixed(0)}ms</span>
                      <span>Conexões: {targetConnections}</span>
                    </div>
                  </div>
                </div>
              );
            })}
          </div>
        </CardContent>
      </Card>
    </div>
  );
};

export default DatabasePage;
