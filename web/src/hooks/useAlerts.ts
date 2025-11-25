import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { api } from "@/lib/api";
import type { AlertRule } from "@/lib/types";
import { toast } from "sonner";

export const useAlertRules = () => {
  return useQuery({
    queryKey: ["alerts", "rules"],
    queryFn: () => api.alerts.list(),
    refetchInterval: 15000,
    retry: 1,
    retryDelay: 1000,
    staleTime: 10000,
  });
};

export const useAlertRule = (id: number) => {
  return useQuery({
    queryKey: ["alerts", "rule", id],
    queryFn: () => api.alerts.get(id),
    enabled: id > 0,
  });
};

export const useCreateAlertRule = () => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (rule: Omit<AlertRule, "id" | "created_at" | "updated_at">) =>
      api.alerts.create(rule),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["alerts", "rules"] });
      toast.success("Regra de alerta criada com sucesso!");
    },
    onError: (error: Error) => {
      toast.error(`Erro ao criar regra: ${error.message}`);
    },
  });
};

export const useUpdateAlertRule = () => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ id, rule }: { id: number; rule: Partial<AlertRule> }) =>
      api.alerts.update(id, rule),
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({ queryKey: ["alerts", "rules"] });
      queryClient.invalidateQueries({ queryKey: ["alerts", "rule", variables.id] });
      toast.success("Regra de alerta atualizada com sucesso!");
    },
    onError: (error: Error) => {
      toast.error(`Erro ao atualizar regra: ${error.message}`);
    },
  });
};

export const useDeleteAlertRule = () => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (id: number) => api.alerts.delete(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["alerts", "rules"] });
      toast.success("Regra de alerta excluída com sucesso!");
    },
    onError: (error: Error) => {
      toast.error(`Erro ao excluir regra: ${error.message}`);
    },
  });
};

