export function getUrlForPath(
  url: URL,
  path: string,
  retainParams = ["features"],
): URL {
  if (!path.startsWith("/")) path = "/" + path;
  const newUrl = new URL(`${url.protocol}//${url.host}${path}`);

  for (const param of retainParams) {
    const value = url.searchParams.get(param);
    if (!value) continue;
    newUrl.searchParams.set(param, value);
  }

  return newUrl;
}

export function getFullUrlForPath(
  url: URL,
  path: string,
  retainParams = ["features"],
): string {
  const newUrl = getUrlForPath(url, path, retainParams);

  if (newUrl.search !== "") {
    return `${newUrl.pathname}${newUrl.search}`;
  }
  return newUrl.pathname;
}
