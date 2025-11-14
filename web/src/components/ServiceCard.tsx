import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import StatusBadge from "./StatusBadge";
import { LucideIcon, ArrowRight } from "lucide-react";
import { Link } from "react-router-dom";

interface ServiceCardProps {
  title: string;
  icon: LucideIcon;
  status: "ok" | "warning" | "critical";
  uptime: string;
  metrics: Array<{ label: string; value: string }>;
  link: string;
}

const ServiceCard = ({ title, icon: Icon, status, uptime, metrics, link }: ServiceCardProps) => {
  return (
    <Card className="overflow-hidden transition-all hover:shadow-xl hover:scale-[1.02]">
      <CardHeader className="bg-gradient-to-br from-card to-secondary">
        <div className="flex items-start justify-between">
          <div className="flex items-center gap-3">
            <div className="rounded-lg bg-primary/10 p-2.5">
              <Icon className="h-6 w-6 text-primary" />
            </div>
            <div>
              <CardTitle className="text-lg">{title}</CardTitle>
              <p className="text-sm text-muted-foreground">Uptime: {uptime}</p>
            </div>
          </div>
          <StatusBadge status={status} />
        </div>
      </CardHeader>
      <CardContent className="pt-6">
        <div className="space-y-3">
          {metrics.map((metric, index) => (
            <div key={index} className="flex items-center justify-between">
              <span className="text-sm text-muted-foreground">{metric.label}</span>
              <span className="text-sm font-medium">{metric.value}</span>
            </div>
          ))}
        </div>
        <Link to={link}>
          <Button variant="outline" className="mt-4 w-full group">
            Ver Detalhes
            <ArrowRight className="ml-2 h-4 w-4 transition-transform group-hover:translate-x-1" />
          </Button>
        </Link>
      </CardContent>
    </Card>
  );
};

export default ServiceCard;
