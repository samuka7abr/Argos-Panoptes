import { Mail, TrendingDown, Clock, AlertCircle } from "lucide-react";
import MetricCard from "@/components/MetricCard";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import StatusBadge from "@/components/StatusBadge";

const SMTP = () => {
  const metrics = [
    { title: "Taxa de Entrega", value: "92", unit: "%", icon: TrendingDown, status: "critical" as const, trend: "down" as const, trendValue: "-5%" },
    { title: "Fila de E-mails", value: "245", icon: Mail, status: "critical" as const },
    { title: "Latência de Envio", value: "1.2", unit: "s", icon: Clock, status: "warning" as const },
    { title: "Taxa de Erro", value: "8", unit: "%", icon: AlertCircle, status: "critical" as const },
  ];

  const queueDetails = [
    { status: "Aguardando", count: 187, color: "bg-status-warning" },
    { status: "Processando", count: 45, color: "bg-primary" },
    { status: "Falhados", count: 13, color: "bg-status-critical" },
  ];

  return (
    <div className="space-y-6">
      <div className="flex items-start justify-between">
        <div className="flex items-center gap-3">
          <div className="rounded-lg bg-primary/10 p-3">
            <Mail className="h-8 w-8 text-primary" />
          </div>
          <div>
            <h2 className="text-3xl font-bold tracking-tight">SMTP Server</h2>
            <p className="text-muted-foreground">Monitoramento de servidor de e-mail</p>
          </div>
        </div>
        <StatusBadge status="critical" />
      </div>

      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
        {metrics.map((metric, index) => (
          <MetricCard key={index} {...metric} />
        ))}
      </div>

      {/* Critical Alert */}
      <Card className="border-status-critical/50 bg-status-critical/5">
        <CardContent className="pt-6">
          <div className="flex items-start gap-3">
            <AlertCircle className="h-5 w-5 text-status-critical mt-0.5" />
            <div>
              <h3 className="font-semibold text-status-critical mb-1">Alerta Crítico</h3>
              <p className="text-sm text-muted-foreground">
                Fila de e-mails está muito alta (245 mensagens). Taxa de entrega caiu para 92%. 
                Verificar configurações do servidor e possíveis bloqueios.
              </p>
            </div>
          </div>
        </CardContent>
      </Card>

      <div className="grid gap-6 md:grid-cols-2">
        <Card>
          <CardHeader>
            <CardTitle>Status da Fila</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="space-y-4">
              {queueDetails.map((item, index) => (
                <div key={index}>
                  <div className="flex items-center justify-between mb-2">
                    <span className="text-sm font-medium">{item.status}</span>
                    <span className="text-sm">{item.count} e-mails</span>
                  </div>
                  <div className="h-2 rounded-full bg-secondary overflow-hidden">
                    <div className={`h-full ${item.color} rounded-full transition-all`} style={{ width: `${(item.count / 245) * 100}%` }} />
                  </div>
                </div>
              ))}
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>Volume de E-mails</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="space-y-4">
              <div className="grid grid-cols-2 gap-4">
                <div>
                  <p className="text-xs text-muted-foreground">Última Hora</p>
                  <p className="text-2xl font-bold">1,234</p>
                </div>
                <div>
                  <p className="text-xs text-muted-foreground">Hoje</p>
                  <p className="text-2xl font-bold">28,567</p>
                </div>
              </div>
              <div className="grid grid-cols-2 gap-4">
                <div>
                  <p className="text-xs text-muted-foreground">Esta Semana</p>
                  <p className="text-2xl font-bold">178K</p>
                </div>
                <div>
                  <p className="text-xs text-muted-foreground">Este Mês</p>
                  <p className="text-2xl font-bold">1.2M</p>
                </div>
              </div>
            </div>
          </CardContent>
        </Card>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>Erros Recentes</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="space-y-3">
            <div className="rounded-lg border border-border p-4 hover:bg-accent/50 transition-colors">
              <div className="flex items-start justify-between mb-2">
                <span className="text-sm font-medium text-status-critical">Timeout</span>
                <span className="text-xs text-muted-foreground">5 min atrás</span>
              </div>
              <p className="text-sm text-muted-foreground">Conexão timeout ao tentar enviar para smtp.provider.com</p>
            </div>
            <div className="rounded-lg border border-border p-4 hover:bg-accent/50 transition-colors">
              <div className="flex items-start justify-between mb-2">
                <span className="text-sm font-medium text-status-critical">Rejeitado</span>
                <span className="text-xs text-muted-foreground">12 min atrás</span>
              </div>
              <p className="text-sm text-muted-foreground">E-mail rejeitado por políticas do servidor de destino</p>
            </div>
          </div>
        </CardContent>
      </Card>
    </div>
  );
};

export default SMTP;
