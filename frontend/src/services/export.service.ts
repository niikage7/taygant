import { httpClient } from "./http-client";
import type { ExportFormat } from "@/types";

/** Результат выгрузки: содержимое файла и имя для сохранения. */
export type ExportFile = {
  blob: Blob;
  filename: string;
};

function fileNameFromDisposition(value: string | null, fallback: string): string {
  if (!value) return fallback;
  const match = /filename\*?=(?:"([^"]+)"|([^;]+))/.exec(value);
  const raw = match?.[1] ?? match?.[2];
  if (!raw) return fallback;
  try {
    return decodeURIComponent(raw.trim());
  } catch {
    return raw.trim();
  }
}

export const exportService = {
  /**
   * Формирует файл выгрузки проекта (реестр задач, диаграмма Ганта, сводка)
   * в выбранном формате и подхватывает имя файла из Content-Disposition.
   */
  async export(projectId: string, format: ExportFormat): Promise<ExportFile> {
    const { blob, contentDisposition } = await httpClient.getBlob(
      `/projects/${projectId}/export`,
      { query: { format } },
    );
    return {
      blob,
      filename: fileNameFromDisposition(
        contentDisposition,
        `taygant-${projectId}.${format}`,
      ),
    };
  },
};