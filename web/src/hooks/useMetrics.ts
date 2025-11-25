import { useQuery } from "@tanstack/react-query";
import { api } from "@/lib/api";

export const useHealth = () => {
  return useQuery({
    queryKey: ["health"],
    queryFn: () => api.health.get(),
    refetchInterval: 10000,
    retry: 1,
    retryDelay: 1000,
    staleTime: 5000,
  });
};

export const useLatestMetrics = () => {
  return useQuery({
    queryKey: ["metrics", "latest"],
    queryFn: () => api.metrics.latest(),
    refetchInterval: 10000,
    retry: 1,
    retryDelay: 1000,
    staleTime: 5000,
  });
};

export const useMetricQuery = (
  service: string,
  target: string,
  metricName: string,
  duration: string,
  enabled = true
) => {
  return useQuery({
    queryKey: ["metrics", "query", service, target, metricName, duration],
    queryFn: () => api.metrics.query(service, target, metricName, duration),
    enabled: enabled && !!service && !!target && !!metricName,
    refetchInterval: 15000,
    retry: 1,
    retryDelay: 1000,
    staleTime: 10000,
  });
};

export const useMetricRange = (
  service: string,
  target: string,
  metricName: string,
  start: string,
  end: string,
  enabled = true
) => {
  return useQuery({
    queryKey: ["metrics", "range", service, target, metricName, start, end],
    queryFn: () => api.metrics.range(service, target, metricName, start, end),
    enabled: enabled && !!service && !!target && !!metricName && !!start && !!end,
  });
};

export const useServices = () => {
  return useQuery({
    queryKey: ["metrics", "services"],
    queryFn: () => api.metrics.services(),
    staleTime: 60000,
  });
};

export const useTargets = (service: string) => {
  return useQuery({
    queryKey: ["metrics", "targets", service],
    queryFn: () => api.metrics.targets(service),
    enabled: !!service,
    staleTime: 60000,
  });
};

