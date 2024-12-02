export interface Problem {
  id: number
  title: string
  difficulty: 1 | 2 | 3 | 4 | 5
  acceptedCount: number
  submissionCount: number
  source: string
  role: 'user' | 'admin'
  status: string
}
