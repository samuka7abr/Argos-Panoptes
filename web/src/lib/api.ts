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

  security: {
    events: (limit?: number) =>
      fetchAPI<{ events: SecurityEvent[]; count: number }>(
        `/api/security/events${limit ? `?limit=${limit}` : ""}`
      ),

    failedLogins: (limit?: number) =>
      fetchAPI<{ by_ip: FailedLoginByIP[]; total: number }>(
        `/api/security/failed-logins${limit ? `?limit=${limit}` : ""}`
      ),

    configChanges: (limit?: number) =>
      fetchAPI<{ changes: ConfigChange[]; count: number }>(
        `/api/security/config-changes${limit ? `?limit=${limit}` : ""}`
      ),

    vulnerabilities: () =>
      fetchAPI<{ vulnerabilities: Vulnerability[]; count: number }>(
        "/api/security/vulnerabilities"
      ),

    stats: () =>
      fetchAPI<{
        failed_logins: number;
        traffic_anomalies: number;
        config_changes: number;
        vulnerabilities: number;
      }>("/api/security/stats"),

    recordEvent: (event: Omit<SecurityEvent, "id" | "created_at">) =>
      fetchAPI<SecurityEvent>("/api/security/record-event", {
        method: "POST",
        body: JSON.stringify(event),
      }),

    recordFailedLogin: (data: {
      ip_address: string;
      username?: string;
      service?: string;
      user_agent?: string;
    }) =>
      fetchAPI<{ status: string }>("/api/security/record-failed-login", {
        method: "POST",
        body: JSON.stringify(data),
      }),
  },
};

