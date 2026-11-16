import type { Metadata } from "next";
import "katex/dist/katex.min.css";
import "./globals.css";

export const metadata: Metadata = {
  title: "BitForge",
  description:
    "Daily practice in systems, AI, and math.",
};

const themeScript = `
try {
  var theme = localStorage.getItem("bf-theme");
  if (theme !== "dark" && theme !== "light") {
    theme = "dark";
  }
  document.documentElement.dataset.theme = theme;
} catch (_) { document.documentElement.dataset.theme = "dark"; }
`;

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="en" data-theme="dark" suppressHydrationWarning>
      <body>
        <script dangerouslySetInnerHTML={{ __html: themeScript }} />
        {children}
      </body>
    </html>
  );
}
