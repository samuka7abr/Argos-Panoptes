import { Server, Database, Globe, Mail, Activity, TrendingUp } from "lucide-react";
import ServiceCard from "@/components/ServiceCard";
import MetricCard from "@/components/MetricCard";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import StatusBadge from "@/components/StatusBadge";

const Dashboard = () => {
  // Mock data - será substituído pela integração com o backend Golang
  const services = [
    {
      title: "Web Server",
      icon: Server,
      status: "ok" as const,
      uptime: "99.9%",
      metrics: [
        { label: "Requests/seg", value: "1,234" },
        { label: "Latência", value: "45ms" },
        { label: "Taxa de Erro", value: "0.01%" },
      ],
      link: "/webserver",
    },
    {
      title: "Database",
      icon: Database,
      status: "warning" as const,
      uptime: "99.5%",
      metrics: [
        { label: "Queries/seg", value: "856" },
        { label: "Conexões Ativas", value: "42" },
        { label: "Pool Size", value: "85%" },
      ],
      link: "/database",
    },
    {
      title: "DNS",
      icon: Globe,
      status: "ok" as const,
      uptime: "100%",
      metrics: [
        { label: "Tempo de Resposta", value: "12ms" },
        { label: "Queries/seg", value: "523" },
        { label: "Taxa de Erro", value: "0%" },
      ],
      link: "/dns",
    },
    {
      title: "SMTP",
      icon: Mail,
      status: "critical" as const,
      uptime: "95.2%",
      metrics: [
        { label: "Taxa de Entrega", value: "92%" },
        { label: "Fila", value: "245" },
        { label: "Latência", value: "1.2s" },
      ],
      link: "/smtp",
    },
  ];

  const alerts = [
    { service: "SMTP", message: "Fila de e-mails elevada detectada", status: "critical" as const, time: "2min atrás" },
    { service: "Database", message: "Pool de conexões acima de 80%", status: "warning" as const, time: "15min atrás" },
  ];

  return (
    <div className="space-y-6">
      {/* Header Stats */}
      <div>
        <h2 className="text-3xl font-bold tracking-tight mb-1">Dashboard de Monitoramento</h2>
        <p className="text-muted-foreground">Visão consolidada de todos os serviços</p>
      </div>

      {/* Key Metrics */}
      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
        <MetricCard title="Serviços Ativos" value="4/4" icon={Activity} status="ok" trend="stable" />
        <MetricCard title="Taxa de Disponibilidade" value="98.7" unit="%" icon={TrendingUp} status="ok" trend="up" trendValue="+0.3%" />
        <MetricCard title="Alertas Ativos" value="2" icon={Activity} status="warning" />
        <MetricCard title="Tempo Médio Resposta" value="42" unit="ms" icon={Activity} status="ok" trend="down" trendValue="-5ms" />
      </div>

      {/* Alerts Section */}
      {alerts.length > 0 && (
        <Card>
          <CardHeader>
            <CardTitle>Alertas Recentes</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="space-y-3">
              {alerts.map((alert, index) => (
                <div key={index} className="flex items-start justify-between rounded-lg border border-border p-4 transition-colors hover:bg-accent/50">
                  <div className="flex-1">
                    <div className="flex items-center gap-2 mb-1">
                      <StatusBadge status={alert.status} />
                      <span className="font-medium">{alert.service}</span>
                    </div>
                    <p className="text-sm text-muted-foreground">{alert.message}</p>
                  </div>
                  <span className="text-xs text-muted-foreground">{alert.time}</span>
                </div>
              ))}
            </div>
          </CardContent>
        </Card>
      )}

      {/* Services Grid */}
      <div>
        <h3 className="text-xl font-semibold mb-4">Serviços Monitorados</h3>
        <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-2">
          {services.map((service, index) => (
            <ServiceCard key={index} {...service} />
          ))}
        </div>
      </div>
    </div>
  );
};

export default Dashboard;
