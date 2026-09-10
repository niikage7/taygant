import { httpClient } from "./http-client";
import type { ExportFormat } from "@/types";

export const exportService = {
  /** Формирует файл выгрузки проекта (реестр задач, диаграмма Ганта, сводка) в выбранном формате. */
  export(projectId: string, format: ExportFormat): Promise<Blob> {
    return httpClient.getBlob(`/projects/${projectId}/export`, {
      query: { format },
    });
  },
};
