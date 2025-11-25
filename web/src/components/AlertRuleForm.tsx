import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import * as z from "zod";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import {
  Form,
  FormControl,
  FormDescription,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from "@/components/ui/form";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Switch } from "@/components/ui/switch";
import { useCreateAlertRule, useUpdateAlertRule } from "@/hooks/useAlerts";
import type { AlertRule } from "@/lib/types";

const alertRuleSchema = z.object({
  name: z.string().min(3, "Nome deve ter no mínimo 3 caracteres"),
  description: z.string().optional(),
  expr: z.string().min(1, "Expressão é obrigatória"),
  service: z.string().optional(),
  target: z.string().optional(),
  for_duration: z.string().min(1, "Duração é obrigatória"),
  severity: z.enum(["info", "warning", "critical"]),
  email_to: z.string().min(1, "Pelo menos um e-mail é necessário"),
  enabled: z.boolean().default(true),
});

type AlertRuleFormData = z.infer<typeof alertRuleSchema>;

interface AlertRuleFormProps {
  initialData?: AlertRule;
  onSuccess?: () => void;
}

const AlertRuleForm = ({ initialData, onSuccess }: AlertRuleFormProps) => {
  const createRule = useCreateAlertRule();
  const updateRule = useUpdateAlertRule();

  const form = useForm<AlertRuleFormData>({
    resolver: zodResolver(alertRuleSchema),
    defaultValues: {
      name: initialData?.name || "",
      description: initialData?.description || "",
      expr: initialData?.expr || "",
      service: initialData?.service || "",
      target: initialData?.target || "",
      for_duration: initialData?.for_duration || "1m",
      severity: initialData?.severity || "warning",
      email_to: initialData?.email_to?.join(", ") || "",
      enabled: initialData?.enabled ?? true,
    },
  });

  const onSubmit = async (data: AlertRuleFormData) => {
    const emailArray = data.email_to
      .split(",")
      .map((email) => email.trim())
      .filter((email) => email.length > 0);

    const ruleData = {
      ...data,
      email_to: emailArray,
    };

    if (initialData?.id) {
      await updateRule.mutateAsync({ id: initialData.id, rule: ruleData });
    } else {
      await createRule.mutateAsync(ruleData);
    }

    onSuccess?.();
  };

  return (
    <Form {...form}>
      <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-6">
        <FormField
          control={form.control}
          name="name"
          render={({ field }) => (
            <FormItem>
              <FormLabel>Nome da Regra</FormLabel>
              <FormControl>
                <Input placeholder="http-down-critical" {...field} />
              </FormControl>
              <FormDescription>Identificador único para a regra</FormDescription>
              <FormMessage />
            </FormItem>
          )}
        />

        <FormField
          control={form.control}
          name="description"
          render={({ field }) => (
            <FormItem>
              <FormLabel>Descrição (Opcional)</FormLabel>
              <FormControl>
                <Textarea
                  placeholder="Alerta quando o serviço HTTP está indisponível"
                  {...field}
                />
              </FormControl>
              <FormMessage />
            </FormItem>
          )}
        />

        <FormField
          control={form.control}
          name="expr"
          render={({ field }) => (
            <FormItem>
              <FormLabel>Expressão</FormLabel>
              <FormControl>
                <Input placeholder="last(1m, http_up) == 0" {...field} />
              </FormControl>
              <FormDescription>
                Suporta: last(duration, metric), avg_over(duration, metric), zscore(duration,
                metric)
              </FormDescription>
              <FormMessage />
            </FormItem>
          )}
        />

        <div className="grid grid-cols-2 gap-4">
          <FormField
            control={form.control}
            name="service"
            render={({ field }) => (
              <FormItem>
                <FormLabel>Serviço (Opcional)</FormLabel>
                <FormControl>
                  <Input placeholder="http" {...field} />
                </FormControl>
                <FormMessage />
              </FormItem>
            )}
          />

          <FormField
            control={form.control}
            name="target"
            render={({ field }) => (
              <FormItem>
                <FormLabel>Target (Opcional)</FormLabel>
                <FormControl>
                  <Input placeholder="example.com" {...field} />
                </FormControl>
                <FormMessage />
              </FormItem>
            )}
          />
        </div>

        <div className="grid grid-cols-2 gap-4">
          <FormField
            control={form.control}
            name="for_duration"
            render={({ field }) => (
              <FormItem>
                <FormLabel>Duração</FormLabel>
                <FormControl>
                  <Input placeholder="2m" {...field} />
                </FormControl>
                <FormDescription>Ex: 1m, 5m, 1h</FormDescription>
                <FormMessage />
              </FormItem>
            )}
          />

          <FormField
            control={form.control}
            name="severity"
            render={({ field }) => (
              <FormItem>
                <FormLabel>Severidade</FormLabel>
                <Select onValueChange={field.onChange} defaultValue={field.value}>
                  <FormControl>
                    <SelectTrigger>
                      <SelectValue placeholder="Selecione a severidade" />
                    </SelectTrigger>
                  </FormControl>
                  <SelectContent>
                    <SelectItem value="info">Info</SelectItem>
                    <SelectItem value="warning">Warning</SelectItem>
                    <SelectItem value="critical">Critical</SelectItem>
                  </SelectContent>
                </Select>
                <FormMessage />
              </FormItem>
            )}
          />
        </div>

        <FormField
          control={form.control}
          name="email_to"
          render={({ field }) => (
            <FormItem>
              <FormLabel>E-mails para Notificação</FormLabel>
              <FormControl>
                <Input placeholder="admin@example.com, ops@example.com" {...field} />
              </FormControl>
              <FormDescription>Separe múltiplos e-mails com vírgula</FormDescription>
              <FormMessage />
            </FormItem>
          )}
        />

        <FormField
          control={form.control}
          name="enabled"
          render={({ field }) => (
            <FormItem className="flex flex-row items-center justify-between rounded-lg border p-4">
              <div className="space-y-0.5">
                <FormLabel className="text-base">Regra Ativa</FormLabel>
                <FormDescription>
                  Regras inativas não dispararão alertas
                </FormDescription>
              </div>
              <FormControl>
                <Switch checked={field.value} onCheckedChange={field.onChange} />
              </FormControl>
            </FormItem>
          )}
        />

        <div className="flex justify-end gap-3">
          <Button
            type="submit"
            disabled={createRule.isPending || updateRule.isPending}
          >
            {createRule.isPending || updateRule.isPending
              ? "Salvando..."
              : initialData
                ? "Atualizar"
                : "Criar"}
          </Button>
        </div>
      </form>
    </Form>
  );
};

export default AlertRuleForm;

