import { Mail, Activity, Clock } from "lucide-react";
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

const SMTP = () => {
  const { data: latestMetrics, isLoading } = useLatestMetrics();
  const smtpMetrics = latestMetrics?.filter((m) => m.service === "smtp");
  const [selectedTarget, setSelectedTarget] = useState<string>("");

  const currentTarget = selectedTarget || smtpMetrics?.[0]?.target || "";

  const { data: handshakeData } = useMetricQuery(
    "smtp",
    currentTarget,
    "smtp_handshake_ms",
    "1h",
    !!currentTarget
  );

  const currentMetrics = smtpMetrics?.find((m) => m.target === currentTarget);

  const isUp = currentMetrics?.metrics.smtp_up === 1;
  const handshakeDuration = currentMetrics?.metrics.smtp_handshake_ms || 0;
  const supportsTLS = currentMetrics?.metrics.smtp_supports_tls === 1;

  const chartData =
    handshakeData?.data.map((d) => ({
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

  if (!smtpMetrics || smtpMetrics.length === 0) {
    return (
      <div className="space-y-6">
        <div className="flex items-center gap-3">
          <div className="rounded-lg bg-primary/10 p-3">
            <Mail className="h-8 w-8 text-primary" />
          </div>
          <div>
            <h2 className="text-3xl font-bold tracking-tight">SMTP</h2>
            <p className="text-muted-foreground">Monitoramento de serviço de e-mail</p>
          </div>
        </div>
        <Card>
          <CardContent className="flex flex-col items-center justify-center py-12">
            <Mail className="h-12 w-12 text-muted-foreground mb-4" />
            <p className="text-lg font-medium">Nenhum serviço SMTP monitorado</p>
            <p className="text-sm text-muted-foreground mt-2">
              Configure o Agent para monitorar servidores SMTP
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
            <Mail className="h-8 w-8 text-primary" />
          </div>
          <div>
            <h2 className="text-3xl font-bold tracking-tight">SMTP</h2>
            <p className="text-muted-foreground">Monitoramento de serviço de e-mail</p>
          </div>
        </div>
        <StatusBadge status={isUp ? "ok" : "critical"} />
      </div>

      {smtpMetrics.length > 1 && (
        <div className="flex items-center gap-3">
          <label className="text-sm font-medium">Target:</label>
          <Select value={currentTarget} onValueChange={setSelectedTarget}>
            <SelectTrigger className="w-[300px]">
              <SelectValue placeholder="Selecione um target" />
            </SelectTrigger>
            <SelectContent>
              {smtpMetrics.map((m) => (
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
          title="Tempo de Handshake"
          value={handshakeDuration.toFixed(0)}
          unit="ms"
          icon={Clock}
          status={
            handshakeDuration < 200 ? "ok" : handshakeDuration < 1000 ? "warning" : "critical"
          }
        />
        <MetricCard
          title="Targets Ativos"
          value={smtpMetrics.length}
          icon={Activity}
          status="ok"
        />
      </div>

      <Card>
        <CardHeader>
          <CardTitle>Tempo de Handshake (última hora)</CardTitle>
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
          <CardTitle>Capacidades do Servidor</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="grid gap-4 md:grid-cols-2">
            <div className="space-y-2">
              <p className="text-sm text-muted-foreground">Suporte TLS/STARTTLS</p>
              <p className="text-2xl font-bold">
                {supportsTLS ? (
                  <span className="text-status-ok">Sim ✓</span>
                ) : (
                  <span className="text-status-warning">Não</span>
                )}
              </p>
            </div>
            <div className="space-y-2">
              <p className="text-sm text-muted-foreground">Status da Conexão</p>
              <p className="text-2xl font-bold">
                {isUp ? (
                  <span className="text-status-ok">Online</span>
                ) : (
                  <span className="text-status-critical">Offline</span>
                )}
              </p>
            </div>
          </div>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>Todos os Servidores SMTP</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="space-y-3">
            {smtpMetrics.map((metric) => {
              const targetUp = metric.metrics.smtp_up === 1;
              const targetDuration = metric.metrics.smtp_handshake_ms || 0;
              const targetTLS = metric.metrics.smtp_supports_tls === 1;

              return (
                <div
                  key={metric.target}
                  className="flex items-center justify-between rounded-lg border p-3"
                >
                  <div className="flex-1">
                    <div className="flex items-center gap-2 mb-1">
                      <StatusBadge status={targetUp ? "ok" : "critical"} />
                      <span className="font-medium">{metric.target}</span>
                      {targetTLS && (
                        <span className="text-xs bg-status-ok/20 text-status-ok px-2 py-0.5 rounded">
                          TLS
                        </span>
                      )}
                    </div>
                    <div className="text-xs text-muted-foreground">
                      Handshake: {targetDuration.toFixed(0)}ms
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

export default SMTP;
