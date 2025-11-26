import { Activity, Clock, TrendingUp, AlertCircle, Server } from "lucide-react";
import MetricCard from "@/components/MetricCard";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import StatusBadge from "@/components/StatusBadge";
import { useLatestMetrics, useMetricQuery, useTargets } from "@/hooks/useMetrics";
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

const WebServer = () => {
  const { data: latestMetrics, isLoading } = useLatestMetrics();
  const { data: targets } = useTargets("web");
  const [selectedTarget, setSelectedTarget] = useState<string>("");

  const httpMetrics = latestMetrics?.filter(
    (m) => m.service === "web"
  );

  const currentTarget = selectedTarget || httpMetrics?.[0]?.target || "";

  const { data: latencyData } = useMetricQuery(
    "web",
    currentTarget,
    "http_latency_ms",
    "1h",
    !!currentTarget
  );

  const currentMetrics = httpMetrics?.find((m) => m.target === currentTarget);

  const isUp = currentMetrics?.metrics.http_up === 1;
  const latency = currentMetrics?.metrics.http_latency_ms || 0;
  const statusCode = currentMetrics?.metrics.http_status_code || 0;

  const chartData =
    latencyData?.data.map((d) => ({
      time: new Date(d.timestamp).toLocaleTimeString("pt-BR", {
        hour: "2-digit",
        minute: "2-digit",
      }),
      latency: d.value,
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

  if (!httpMetrics || httpMetrics.length === 0) {
    return (
      <div className="space-y-6">
        <div className="flex items-center gap-3">
          <div className="rounded-lg bg-primary/10 p-3">
            <Server className="h-8 w-8 text-primary" />
          </div>
          <div>
            <h2 className="text-3xl font-bold tracking-tight">Web Server (HTTP/HTTPS)</h2>
            <p className="text-muted-foreground">Monitoramento de servidor web e APIs</p>
          </div>
        </div>
        <Card>
          <CardContent className="flex flex-col items-center justify-center py-12">
            <Server className="h-12 w-12 text-muted-foreground mb-4" />
            <p className="text-lg font-medium">Nenhum serviço HTTP monitorado</p>
            <p className="text-sm text-muted-foreground mt-2">
              Configure o Agent para monitorar serviços HTTP/HTTPS
            </p>
          </CardContent>
        </Card>
      </div>
    );
  }

  return (
    <div className="space-y-6">
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
        <StatusBadge status={isUp ? "ok" : "critical"} />
      </div>

      {httpMetrics.length > 1 && (
        <div className="flex items-center gap-3">
          <label className="text-sm font-medium">Target:</label>
          <Select value={currentTarget} onValueChange={setSelectedTarget}>
            <SelectTrigger className="w-[300px]">
              <SelectValue placeholder="Selecione um target" />
            </SelectTrigger>
            <SelectContent>
              {httpMetrics.map((m) => (
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
          trend="stable"
        />
        <MetricCard
          title="Latência"
          value={latency.toFixed(0)}
          unit="ms"
          icon={Clock}
          status={latency < 100 ? "ok" : latency < 500 ? "warning" : "critical"}
        />
        <MetricCard
          title="Status Code"
          value={statusCode}
          icon={TrendingUp}
          status={statusCode >= 200 && statusCode < 300 ? "ok" : "warning"}
        />
        <MetricCard
          title="Targets Ativos"
          value={httpMetrics.length}
          icon={AlertCircle}
          status="ok"
        />
      </div>

      <Card>
        <CardHeader>
          <CardTitle>Latência (última hora)</CardTitle>
        </CardHeader>
        <CardContent>
          {chartData.length > 0 ? (
            <ResponsiveContainer width="100%" height={300}>
              <LineChart data={chartData}>
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
            <div className="flex items-center justify-center h-[300px] text-muted-foreground">
              Sem dados disponíveis
            </div>
          )}
        </CardContent>
      </Card>

      <div className="grid gap-6 md:grid-cols-2">
        <Card>
          <CardHeader>
            <CardTitle>Todos os Targets</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="space-y-3">
              {httpMetrics.map((metric) => {
                const targetUp = metric.metrics.http_up === 1;
                const targetLatency = metric.metrics.http_latency_ms || 0;
                const targetStatus = metric.metrics.http_status_code || 0;

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
                        <span>Status: {targetStatus}</span>
                      </div>
                    </div>
                  </div>
                );
              })}
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>Detalhes do Target Atual</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="space-y-4">
              <div>
                <p className="text-sm text-muted-foreground">URL</p>
                <p className="text-lg font-semibold">{currentTarget}</p>
              </div>
              <div>
                <p className="text-sm text-muted-foreground">Status</p>
                <p className="text-lg font-semibold">{isUp ? "Online" : "Offline"}</p>
              </div>
              <div>
                <p className="text-sm text-muted-foreground">Última Verificação</p>
                <p className="text-lg font-semibold">
                  {currentMetrics
                    ? new Date(currentMetrics.timestamp).toLocaleString("pt-BR")
                    : "-"}
                </p>
              </div>
            </div>
          </CardContent>
        </Card>
      </div>
    </div>
  );
};

export default WebServer;
