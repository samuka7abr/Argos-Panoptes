import type {
  QueryRangeResponse,
  LatestMetric,
  AlertRule,
  AlertRulesResponse,
  HealthResponse,
} from "./types";

const API_BASE_URL = import.meta.env.VITE_API_URL || "";
const REQUEST_TIMEOUT = 5000;

class ApiError extends Error {
  constructor(
    public status: number,
    message: string
  ) {
    super(message);
    this.name = "ApiError";
  }
}

async function fetchAPI<T>(endpoint: string, options?: RequestInit): Promise<T> {
  const controller = new AbortController();
  const timeoutId = setTimeout(() => controller.abort(), REQUEST_TIMEOUT);

  try {
    const url = `${API_BASE_URL}${endpoint}`;
    const response = await fetch(url, {
      ...options,
      headers: {
        "Content-Type": "application/json",
        ...options?.headers,
      },
      signal: controller.signal,
    });

    clearTimeout(timeoutId);

    if (!response.ok) {
      throw new ApiError(response.status, `API Error: ${response.statusText}`);
    }

    return response.json();
  } catch (error) {
    clearTimeout(timeoutId);
    
    if (error instanceof Error && error.name === "AbortError") {
      throw new ApiError(408, "Tempo de requisição esgotado");
    }
    
    throw error;
  }
}

export const api = {
  health: {
    get: () => fetchAPI<HealthResponse>("/health"),
  },

  metrics: {
    latest: () => fetchAPI<LatestMetric[]>("/api/metrics/latest"),

    query: (service: string, target: string, metricName: string, duration: string) =>
      fetchAPI<QueryRangeResponse>(
        `/api/metrics/query?service=${service}&target=${target}&metric_name=${metricName}&duration=${duration}`
      ),

    range: (service: string, target: string, metricName: string, start: string, end: string) =>
      fetchAPI<QueryRangeResponse>(
        `/api/metrics/range?service=${service}&target=${target}&metric_name=${metricName}&start=${start}&end=${end}`
      ),

    services: () => fetchAPI<{ services: string[] }>("/api/metrics/services"),

    targets: (service: string) =>
      fetchAPI<{ targets: string[] }>(`/api/metrics/targets?service=${service}`),
  },

  alerts: {
    list: () => fetchAPI<AlertRulesResponse>("/api/alert-rules"),

    get: (id: number) => fetchAPI<AlertRule>(`/api/alert-rules/${id}`),

    create: (rule: Omit<AlertRule, "id" | "created_at" | "updated_at">) =>
      fetchAPI<AlertRule>("/api/alert-rules", {
        method: "POST",
        body: JSON.stringify(rule),
      }),

    update: (id: number, rule: Partial<AlertRule>) =>
      fetchAPI<AlertRule>(`/api/alert-rules/${id}`, {
        method: "PUT",
        body: JSON.stringify(rule),
      }),

    delete: (id: number) =>
      fetchAPI<{ message: string }>(`/api/alert-rules/${id}`, {
        method: "DELETE",
      }),
  },
};

