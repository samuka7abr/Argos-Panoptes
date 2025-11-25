import { useState } from "react";
import { Bell, Plus, Pencil, Trash2, Power, PowerOff } from "lucide-react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import ErrorState from "@/components/ErrorState";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from "@/components/ui/alert-dialog";
import { useAlertRules, useDeleteAlertRule, useUpdateAlertRule } from "@/hooks/useAlerts";
import AlertRuleForm from "@/components/AlertRuleForm";
import type { AlertRule } from "@/lib/types";

const Alerts = () => {
  const { data, isLoading, error, refetch } = useAlertRules();
  const deleteRule = useDeleteAlertRule();
  const updateRule = useUpdateAlertRule();
  const [editingRule, setEditingRule] = useState<AlertRule | null>(null);
  const [deletingRuleId, setDeletingRuleId] = useState<number | null>(null);
  const [isCreateDialogOpen, setIsCreateDialogOpen] = useState(false);

  const handleToggleEnabled = (rule: AlertRule) => {
    if (!rule.id) return;
    updateRule.mutate({
      id: rule.id,
      rule: { enabled: !rule.enabled },
    });
  };

  const handleDelete = () => {
    if (deletingRuleId) {
      deleteRule.mutate(deletingRuleId);
      setDeletingRuleId(null);
    }
  };

  const getSeverityColor = (severity: string) => {
    switch (severity) {
      case "critical":
        return "bg-status-critical text-white";
      case "warning":
        return "bg-status-warning text-black";
      case "info":
        return "bg-blue-500 text-white";
      default:
        return "bg-secondary";
    }
  };

  if (error && !isLoading) {
    return (
      <div className="space-y-6">
        <div className="flex items-center gap-3">
          <div className="rounded-lg bg-primary/10 p-3">
            <Bell className="h-8 w-8 text-primary" />
          </div>
          <div>
            <h2 className="text-3xl font-bold tracking-tight">Gerenciamento de Alertas</h2>
            <p className="text-muted-foreground">Configure regras de alerta para monitoramento</p>
          </div>
        </div>
        <ErrorState onRetry={() => refetch()} />
      </div>
    );
  }

  if (isLoading) {
    return (
      <div className="flex items-center justify-center min-h-[400px]">
        <div className="text-center">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-primary mx-auto" />
          <p className="mt-4 text-muted-foreground">Carregando alertas...</p>
        </div>
      </div>
    );
  }

  const rules = data?.rules || [];

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-3">
          <div className="rounded-lg bg-primary/10 p-3">
            <Bell className="h-8 w-8 text-primary" />
          </div>
          <div>
            <h2 className="text-3xl font-bold tracking-tight">Gerenciamento de Alertas</h2>
            <p className="text-muted-foreground">Configure regras de alerta para monitoramento</p>
          </div>
        </div>

        <Dialog open={isCreateDialogOpen} onOpenChange={setIsCreateDialogOpen}>
          <DialogTrigger asChild>
            <Button>
              <Plus className="mr-2 h-4 w-4" />
              Nova Regra
            </Button>
          </DialogTrigger>
          <DialogContent className="max-w-2xl max-h-[90vh] overflow-y-auto">
            <DialogHeader>
              <DialogTitle>Criar Nova Regra de Alerta</DialogTitle>
            </DialogHeader>
            <AlertRuleForm onSuccess={() => setIsCreateDialogOpen(false)} />
          </DialogContent>
        </Dialog>
      </div>

      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
        <Card>
          <CardHeader className="pb-3">
            <CardTitle className="text-sm font-medium text-muted-foreground">
              Total de Regras
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{data?.count || 0}</div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="pb-3">
            <CardTitle className="text-sm font-medium text-muted-foreground">
              Regras Ativas
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">
              {rules.filter((r) => r.enabled).length}
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="pb-3">
            <CardTitle className="text-sm font-medium text-muted-foreground">
              Regras Críticas
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">
              {rules.filter((r) => r.severity === "critical").length}
            </div>
          </CardContent>
        </Card>
      </div>

      <div className="space-y-4">
        {rules.length === 0 ? (
          <Card>
            <CardContent className="flex flex-col items-center justify-center py-12">
              <Bell className="h-12 w-12 text-muted-foreground mb-4" />
              <p className="text-lg font-medium mb-2">Nenhuma regra configurada</p>
              <p className="text-sm text-muted-foreground mb-4">
                Comece criando sua primeira regra de alerta
              </p>
              <Button onClick={() => setIsCreateDialogOpen(true)}>
                <Plus className="mr-2 h-4 w-4" />
                Criar Primeira Regra
              </Button>
            </CardContent>
          </Card>
        ) : (
          rules.map((rule) => (
            <Card key={rule.id} className={!rule.enabled ? "opacity-60" : ""}>
              <CardContent className="pt-6">
                <div className="flex items-start justify-between">
                  <div className="flex-1">
                    <div className="flex items-center gap-2 mb-2">
                      <h3 className="text-lg font-semibold">{rule.name}</h3>
                      <Badge className={getSeverityColor(rule.severity)}>
                        {rule.severity.toUpperCase()}
                      </Badge>
                      {rule.enabled ? (
                        <Badge variant="outline" className="text-status-ok border-status-ok">
                          <Power className="h-3 w-3 mr-1" />
                          Ativa
                        </Badge>
                      ) : (
                        <Badge variant="outline" className="text-muted-foreground">
                          <PowerOff className="h-3 w-3 mr-1" />
                          Inativa
                        </Badge>
                      )}
                    </div>

                    {rule.description && (
                      <p className="text-sm text-muted-foreground mb-3">{rule.description}</p>
                    )}

                    <div className="grid grid-cols-2 gap-x-6 gap-y-2 text-sm">
                      <div>
                        <span className="text-muted-foreground">Expressão:</span>
                        <code className="ml-2 px-2 py-1 bg-secondary rounded text-xs">
                          {rule.expr}
                        </code>
                      </div>
                      <div>
                        <span className="text-muted-foreground">Duração:</span>
                        <span className="ml-2 font-medium">{rule.for_duration}</span>
                      </div>
                      {rule.service && (
                        <div>
                          <span className="text-muted-foreground">Serviço:</span>
                          <span className="ml-2 font-medium">{rule.service}</span>
                        </div>
                      )}
                      {rule.target && (
                        <div>
                          <span className="text-muted-foreground">Target:</span>
                          <span className="ml-2 font-medium">{rule.target}</span>
                        </div>
                      )}
                      <div className="col-span-2">
                        <span className="text-muted-foreground">E-mails:</span>
                        <span className="ml-2 font-medium">{rule.email_to.join(", ")}</span>
                      </div>
                    </div>
                  </div>

                  <div className="flex gap-2 ml-4">
                    <Button
                      variant="outline"
                      size="sm"
                      onClick={() => handleToggleEnabled(rule)}
                    >
                      {rule.enabled ? (
                        <PowerOff className="h-4 w-4" />
                      ) : (
                        <Power className="h-4 w-4" />
                      )}
                    </Button>

                    <Dialog
                      open={editingRule?.id === rule.id}
                      onOpenChange={(open) => !open && setEditingRule(null)}
                    >
                      <DialogTrigger asChild>
                        <Button
                          variant="outline"
                          size="sm"
                          onClick={() => setEditingRule(rule)}
                        >
                          <Pencil className="h-4 w-4" />
                        </Button>
                      </DialogTrigger>
                      <DialogContent className="max-w-2xl max-h-[90vh] overflow-y-auto">
                        <DialogHeader>
                          <DialogTitle>Editar Regra de Alerta</DialogTitle>
                        </DialogHeader>
                        <AlertRuleForm
                          initialData={editingRule || undefined}
                          onSuccess={() => setEditingRule(null)}
                        />
                      </DialogContent>
                    </Dialog>

                    <Button
                      variant="outline"
                      size="sm"
                      onClick={() => setDeletingRuleId(rule.id || null)}
                    >
                      <Trash2 className="h-4 w-4 text-destructive" />
                    </Button>
                  </div>
                </div>
              </CardContent>
            </Card>
          ))
        )}
      </div>

      <AlertDialog open={!!deletingRuleId} onOpenChange={() => setDeletingRuleId(null)}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Confirmar exclusão</AlertDialogTitle>
            <AlertDialogDescription>
              Tem certeza que deseja excluir esta regra de alerta? Esta ação não pode ser
              desfeita.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>Cancelar</AlertDialogCancel>
            <AlertDialogAction onClick={handleDelete}>Excluir</AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </div>
  );
};

export default Alerts;

