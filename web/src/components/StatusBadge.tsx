import { cn } from "@/lib/utils";
import { CheckCircle, AlertTriangle, XCircle } from "lucide-react";

type StatusType = "ok" | "warning" | "critical";

interface StatusBadgeProps {
  status: StatusType;
  className?: string;
}

const StatusBadge = ({ status, className }: StatusBadgeProps) => {
  const config = {
    ok: {
      icon: CheckCircle,
      label: "OK",
      color: "bg-status-ok/20 text-status-ok border-status-ok/50",
    },
    warning: {
      icon: AlertTriangle,
      label: "Atenção",
      color: "bg-status-warning/20 text-status-warning border-status-warning/50",
    },
    critical: {
      icon: XCircle,
      label: "Crítico",
      color: "bg-status-critical/20 text-status-critical border-status-critical/50",
    },
  };

  const { icon: Icon, label, color } = config[status];

  return (
    <div className={cn("inline-flex items-center gap-1.5 rounded-full border px-3 py-1 text-xs font-medium", color, className)}>
      <Icon className="h-3.5 w-3.5" />
      {label}
    </div>
  );
};

export default StatusBadge;
