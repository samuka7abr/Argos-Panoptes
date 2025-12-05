import { Globe, Activity, Clock } from "lucide-react";
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

const DNS = () => {
  const { data: latestMetrics, isLoading } = useLatestMetrics();
  const dnsMetrics = latestMetrics?.filter((m) => m.service === "dns");
  const [selectedTarget, setSelectedTarget] = useState<string>("");

  const currentTarget = selectedTarget || dnsMetrics?.[0]?.target || "";

  const { data: lookupData } = useMetricQuery(
    "dns",
    currentTarget,
    "dns_lookup_ms",
    "1h",
    !!currentTarget
  );

  const currentMetrics = dnsMetrics?.find((m) => m.target === currentTarget);

  const isUp = currentMetrics?.metrics.dns_up === 1;
  const lookupDuration = currentMetrics?.metrics.dns_lookup_ms || 0;

  const chartData =
    lookupData?.data.map((d) => ({
      time: new Date(d.timestamp).toLocaleTimeString("pt-BR", {
        hour: "2-digit",
        minute: "2-digit",
      }),
      duration: d.value,
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

  if (!dnsMetrics || dnsMetrics.length === 0) {
    return (
      <div className="space-y-6">
        <div className="flex items-center gap-3">
          <div className="rounded-lg bg-primary/10 p-3">
            <Globe className="h-8 w-8 text-primary" />
          </div>
          <div>
            <h2 className="text-3xl font-bold tracking-tight">DNS</h2>
            <p className="text-muted-foreground">Monitoramento de serviço DNS</p>
          </div>
        </div>
        <Card>
          <CardContent className="flex flex-col items-center justify-center py-12">
            <Globe className="h-12 w-12 text-muted-foreground mb-4" />
            <p className="text-lg font-medium">Nenhum serviço DNS monitorado</p>
            <p className="text-sm text-muted-foreground mt-2">
              Configure o Agent para monitorar servidores DNS
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
            <Globe className="h-8 w-8 text-primary" />
          </div>
          <div>
            <h2 className="text-3xl font-bold tracking-tight">DNS</h2>
            <p className="text-muted-foreground">Monitoramento de serviço DNS</p>
          </div>
        </div>
        <StatusBadge status={isUp ? "ok" : "critical"} />
      </div>

      {dnsMetrics.length > 1 && (
        <div className="flex items-center gap-3">
          <label className="text-sm font-medium">Target:</label>
          <Select value={currentTarget} onValueChange={setSelectedTarget}>
            <SelectTrigger className="w-[300px]">
              <SelectValue placeholder="Selecione um target" />
            </SelectTrigger>
            <SelectContent>
              {dnsMetrics.map((m) => (
                <SelectItem key={m.target} value={m.target}>
                  {m.target}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>
      )}

      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
        <MetricCard
          title="Disponibilidade"
          value={isUp ? "100" : "0"}
          unit="%"
          icon={Activity}
          status={isUp ? "ok" : "critical"}
        />
        <MetricCard
          title="Tempo de Lookup"
          value={lookupDuration.toFixed(0)}
          unit="ms"
          icon={Clock}
          status={lookupDuration < 50 ? "ok" : lookupDuration < 200 ? "warning" : "critical"}
        />
        <MetricCard
          title="Targets Ativos"
          value={dnsMetrics.length}
          icon={Activity}
          status="ok"
        />
      </div>

      <Card>
        <CardHeader>
          <CardTitle>Tempo de Lookup (última hora)</CardTitle>
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
                  dataKey="duration"
                  stroke="hsl(var(--primary))"
                  name="Duração (ms)"
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

      <Card>
        <CardHeader>
          <CardTitle>Todos os Targets DNS</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="space-y-3">
            {dnsMetrics.map((metric) => {
              const targetUp = metric.metrics.dns_up === 1;
              const targetDuration = metric.metrics.dns_lookup_ms || 0;

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
                    <div className="text-xs text-muted-foreground">
                      Tempo de lookup: {targetDuration.toFixed(0)}ms
                    </div>
                  </div>
                  <div className="text-right">
                    <p className="text-sm text-muted-foreground">Última verificação</p>
                    <p className="text-xs">
                      {new Date(metric.timestamp).toLocaleTimeString("pt-BR")}
                    </p>
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

export default DNS;
