let base = document.querySelector("html>head>base")?.getAttribute("href");
if (base === null || base === undefined) {
  base = "/";
}
if (!base.endsWith("/")) {
  base = base + "/";
}
export const baseURL: string = base;
