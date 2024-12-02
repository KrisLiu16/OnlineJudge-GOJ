export interface Submission {
  ID: number
  UserID: number
  ProblemID: string
  problemTitle: string
  Language: string
  Code: string
  Status: string
  TimeUsed: number
  MemoryUsed: number
  ErrorInfo: string
  SubmitTime: string
  JudgeTime: string | null
  userAvatar: string
  username: string
}

export interface SubmissionResponse {
  submissions: Submission[]
  total: number
}

export interface Problem {
  id: string
  title: string
  difficulty: string
  solvedTime: string
}

export interface Contest {
  id: string
  title: string
  startTime: string
  endTime: string
  rank: number
  score: number
  totalParticipants: number
}

export interface SubmissionUpdate {
  ID: number
  Status: string
  TimeUsed: number
  MemoryUsed: number
  ErrorInfo: string
}

export interface WebSocketMessage {
  type: string
  submission: SubmissionUpdate
}
