import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { LucideIcon } from "lucide-react";
import { cn } from "@/lib/utils";

interface MetricCardProps {
  title: string;
  value: string | number;
  unit?: string;
  icon: LucideIcon;
  trend?: "up" | "down" | "stable";
  trendValue?: string;
  status?: "ok" | "warning" | "critical";
  className?: string;
}

const MetricCard = ({ title, value, unit, icon: Icon, trend, trendValue, status = "ok", className }: MetricCardProps) => {
  const statusColors = {
    ok: "border-status-ok/30",
    warning: "border-status-warning/30",
    critical: "border-status-critical/30",
  };

  return (
    <Card className={cn("border-l-4 transition-all hover:shadow-lg", statusColors[status], className)}>
      <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
        <CardTitle className="text-sm font-medium text-muted-foreground">{title}</CardTitle>
        <Icon className="h-4 w-4 text-muted-foreground" />
      </CardHeader>
      <CardContent>
        <div className="flex items-baseline gap-1">
          <div className="text-2xl font-bold">
            {value}
            {unit && <span className="text-base font-normal text-muted-foreground ml-1">{unit}</span>}
          </div>
        </div>
        {trend && trendValue && (
          <p className={cn("text-xs mt-1", trend === "up" ? "text-status-ok" : trend === "down" ? "text-status-critical" : "text-muted-foreground")}>
            {trend === "up" ? "↑" : trend === "down" ? "↓" : "→"} {trendValue}
          </p>
        )}
      </CardContent>
    </Card>
  );
};

export default MetricCard;
