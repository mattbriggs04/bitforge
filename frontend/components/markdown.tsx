import ReactMarkdown, { type Components } from "react-markdown";
import remarkMath from "remark-math";
import rehypeKatex, { type Options as KatexOptions } from "rehype-katex";

const mathOptions: KatexOptions = {
  output: "htmlAndMathml",
  trust: false,
  strict: "warn",
  maxExpand: 1000,
  maxSize: 20,
  errorColor: "currentColor",
};

// Labels and sample notes need inline markup without a paragraph inside a span.
const mathComponents: Components = {
  span: ({ node, ...props }) => node?.tagName === "span" && props.className?.split(" ").includes("katex-display")
    ? <span {...props} tabIndex={0} role="group" aria-label="Equation" />
    : <span {...props} />,
};

const inlineComponents: Components = {
  ...mathComponents,
  p: ({ children }) => <span>{children}</span>,
};

export function Markdown({ children, inline = false }: { children: string; inline?: boolean }) {
  return (
    <ReactMarkdown
      remarkPlugins={[remarkMath]}
      rehypePlugins={[[rehypeKatex, mathOptions]]}
      components={inline ? inlineComponents : mathComponents}
      skipHtml
    >
      {children}
    </ReactMarkdown>
  );
}
