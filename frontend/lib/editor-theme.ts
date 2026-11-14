import type { BeforeMount } from "@monaco-editor/react";

export const configureEditor: BeforeMount = (monaco) => {
  const style = getComputedStyle(document.documentElement);
  const graphite = style.getPropertyValue("--graphite").trim();
  const chalk = style.getPropertyValue("--chalk").trim();
  const acid = style.getPropertyValue("--acid").trim();
  for (const dark of [true, false]) {
    const foreground = dark ? chalk : graphite;
    const background = dark ? graphite : chalk;
    const accent = dark ? acid : graphite;
    monaco.editor.defineTheme(`bitforge-${dark ? "dark" : "light"}`, {
      base: dark ? "vs-dark" : "vs",
      inherit: false,
      rules: [
        { token: "", foreground: foreground.slice(1), background: background.slice(1) },
        { token: "comment", foreground: foreground.slice(1), fontStyle: "italic" },
        { token: "keyword", foreground: accent.slice(1) },
        { token: "number", foreground: accent.slice(1) },
        { token: "string", foreground: foreground.slice(1) },
        { token: "type", foreground: foreground.slice(1) },
      ],
      colors: {
        "editor.background": background,
        "editor.foreground": foreground,
        "editorLineNumber.foreground": foreground + "B3",
        "editorLineNumber.activeForeground": foreground,
        "editorCursor.foreground": dark ? acid : graphite,
        "editor.selectionBackground": dark ? acid + "25" : graphite + "20",
        "editor.inactiveSelectionBackground": foreground + "15",
        "editor.lineHighlightBackground": foreground + "06",
        "editorIndentGuide.background1": foreground + "12",
        "editorIndentGuide.activeBackground1": foreground + "30",
        "editorWhitespace.foreground": foreground + "20",
        "editorWidget.background": background,
        "editorWidget.foreground": foreground,
        "editorWidget.border": foreground + "30",
        "editorSuggestWidget.background": background,
        "editorSuggestWidget.foreground": foreground,
        "editorSuggestWidget.border": foreground + "30",
        "editorSuggestWidget.selectedBackground": foreground + "15",
        "list.highlightForeground": accent,
        ...Object.fromEntries([1,2,3,4,5,6].flatMap((i) => [
          [`editorBracketHighlight.foreground${i}`, foreground],
          [`editorBracketPairGuide.background${i}`, foreground + "15"],
          [`editorBracketPairGuide.activeBackground${i}`, foreground + "30"],
        ])),
        "editorBracketHighlight.unexpectedBracket.foreground": accent,
        "editorError.foreground": foreground,
        "editorWarning.foreground": foreground,
        "editorInfo.foreground": foreground,
        "focusBorder": accent,
        "scrollbarSlider.background": foreground + "16",
        "scrollbarSlider.hoverBackground": foreground + "30",
      },
    });
  }
};
