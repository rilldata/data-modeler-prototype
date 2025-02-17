import type { AxiosResponse } from "axios";
import axios from "axios";
import fs from "fs";
import path from "path";
import { fileURLToPath } from "url";

export async function generateEmbed(
  resourceId: string,
  serviceToken: string,
): Promise<void> {
  try {
    const response: AxiosResponse<{ iframeSrc: string }> = await axios.post(
      "http://localhost:8080/v1/organizations/e2e/projects/openrtb/iframe",
      {
        resource: resourceId,
        navigation: true,
      },
      {
        headers: {
          Authorization: `Bearer ${serviceToken}`,
          "Content-Type": "application/json",
        },
      },
    );

    const iframeSrc = response.data.iframeSrc;
    if (!iframeSrc) {
      throw new Error("Invalid response: iframeSrc not found");
    }

    const htmlContent = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Iframe Example</title>
</head>
<body>
    <iframe id="rill-frame" src="${iframeSrc}" height="600px" width="100%"></iframe>
    <script>
        window.addEventListener('message', (event) => {
            console.log(event.data);
        });
    </script>
</body>
</html>`;

    const __dirname = path.dirname(fileURLToPath(import.meta.url));
    const outputPath = path.join(__dirname, "..", "embed.html");
    console.log(outputPath);
    fs.writeFileSync(outputPath, htmlContent, "utf8");
  } catch (error: unknown) {
    if (error instanceof Error) {
      console.error("Error fetching iframe or saving file:", error.message);
    } else {
      console.error("An unknown error occurred:", error);
    }
  }
}
