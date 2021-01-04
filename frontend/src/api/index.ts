import axios from "axios";

export interface Recording {
  stationId: string;
  start: Date;
  end: Date;
  title: string;
}

export interface Status {
  status: string;
  downloadProgress: number;
}

const baseURL = (function() {
  let u: string = process.env.VUE_APP_API_ENDPOINT;
  if (!u.endsWith("/")) {
    u = u + "/";
  }
  let pathname = window.location.pathname;
  if (pathname.startsWith("/")) {
    pathname = pathname.substring(1);
  }
  if (pathname.length > 0) {
    u = u + pathname;
  }
  return u;
})();

const a = axios.create({
  baseURL,
});

class Api {
  async getRecordings(): Promise<Recording[]> {
    const d = (await a.get("/recordings/")).data as any[];
    return d.map((v) => {
      v.start = new Date(v.start);
      v.end = new Date(v.end);
      return v;
    });
  }
  async record(stationId: string, start: string): Promise<void> {
    a.post("recordings/record", {
      stationId,
      start,
    });
  }
  async getStatus(stationId: string, start: string): Promise<Status> {
    const data = (await a.get(`recordings/recording/${stationId}/${start}`))
      .data;
    return data.status;
  }
  getPlaylistUrl(stationID: string, start: string): string {
    return `${baseURL}recordings/recording/${stationID}/${start}/playlist`;
  }
}

export const api = new Api();
