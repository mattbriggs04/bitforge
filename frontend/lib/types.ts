export type PracticeProfile = { role: string; level: string; focus: string };
export type InterviewContract = { role: string; preset: string; level: string; targetMinutes: number; objectives: string[] };
export type ProblemSummary = {
 practice: PracticeProfile;
 targetMinutes: number;
 objectives: string[];
  id: string;
  slug: string;
  title: string;
  difficulty: "easy" | "medium" | "hard" | string;
  category: string;
  problemType: string;
  shortDescription: string;
  tags: string[];
  completed?: boolean;
};

export type ProblemSample = {
  name: string;
  input: string;
  expected: string;
  explanation: string;
  sortOrder: number;
};

export type LanguageTemplate = {
  language: string;
  starterCode: string;
  notes: string;
};

export type ProblemAsset = {
  assetType: string;
  name: string;
  mimeType: string;
  content: string;
  metadata: Record<string, unknown>;
};

export type ProblemMetadata = Record<string, unknown> & {
  solution?: unknown;
  solutionMarkdown?: string;
  solutionCode?: string;
  solutionLanguage?: string;
};

export type ProblemDetail = ProblemSummary & {
 hints: { nudge: string; approach: string };
  statementMarkdown: string;
  constraintsMarkdown: string;
  samples: ProblemSample[];
  languageTemplates: LanguageTemplate[];
  assets: ProblemAsset[];
  metadata: ProblemMetadata;
  interviews?: InterviewSummary[];
};

export type SubmissionCaseResult = {
  caseName: string;
  isHidden: boolean;
  status: "passed" | "failed" | "error" | "skipped" | string;
  message: string;
  executionMs: number;
  sortOrder: number;
};

export type Submission = {
  id: string;
  problemId: string;
  problemSlug: string;
  userId: string;
  language: string;
  mode: "run" | "submit" | string;
  status: "queued" | "running" | "completed" | "failed" | string;
  verdict:
    | "pending"
    | "accepted"
    | "wrong_answer"
    | "compile_error"
    | "runtime_error"
    | "system_error"
    | string;
  score: number;
  totalTests: number;
  passedTests: number;
  compileOutput?: string;
  runtimeOutput?: string;
  errorMessage?: string;
  queuedAt: string;
  startedAt?: string;
  completedAt?: string;
  results?: SubmissionCaseResult[];
};

export type APIError = {
  error: string;
};

export type InterviewSummary = {
 contract: InterviewContract;
  slug: string;
  title: string;
  category: string;
  description: string;
  questionCount: number;
  targetMinutes: number;
};

export type InterviewDetail = InterviewSummary & {
  problems: ProblemSummary[];
};

export type QuestionMix = { format: string; focus: string; count: number };
export type PromptCatalog = {
 roles: { id: string; label: string; category: string; language: string; emphasis: string }[];
 presets: { id: string; label: string; targetMinutes: number }[];
 formats: string[];
 focuses: string[];
 defaults: Record<string, Record<string, QuestionMix[]>>;
};
export type GenerationConfig = {
 schemaVersion: number; kind: string; role: string; preset: string; category: string;
 format: string; language: string; difficulty: string; level: string; topic: string;
 targetMinutes: number; seed: number; avoid: string[]; objectives: string[]; mix: QuestionMix[];
};
export type CompiledPrompt = { config: GenerationConfig; prompt: string };
