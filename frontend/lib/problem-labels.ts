const labels: Record<string, string> = {
  ai_ml: "AI / ML",
  embedded: "Embedded",
  computer_vision: "Computer vision",
  computer_architecture: "Architecture",
  operating_systems: "Operating systems",
  single_choice: "Single choice",
  multiple_select: "Multiple select",
  numeric: "Numerical",
  coding: "Code",
  applied: "Applied",
  fundamentals: "Fundamentals",
  reasoning: "Reasoning",
  daily: "Daily practice",
  interview: "Technical interview",
  exam: "Exam practice",
};

export function problemLabel(value: string): string {
  return labels[value] ?? value.replaceAll("_", " ");
}
