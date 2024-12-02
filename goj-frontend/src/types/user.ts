export interface User {
  id: number
  username: string
  email: string
  avatar?: string
  bio?: string
  role: string
  createdAt: string
  updatedAt: string
  submissions?: number
  acceptedProblems?: number
  rating?: number
}

export interface LoginData {
  email: string
  password: string
}

export interface RegisterData {
  username: string
  email: string
  password: string
  avatar?: string
}

export interface LoginResponse {
  token: string
  user: User
}

export interface ApiResponse<T> {
  code: number
  message?: string
  data: T
}

export interface PasswordUpdateData {
  oldPassword: string
  newPassword: string
}

export interface PasswordUpdateResponse {
  message: string
}

export interface UserApi {
  login: (data: LoginData) => Promise<ApiResponse<LoginResponse>>
  register: (data: RegisterData) => Promise<ApiResponse<null>>
  getProfile: () => Promise<ApiResponse<User>>
  updateProfile: (data: Partial<User>) => Promise<ApiResponse<User>>
  updateAvatar: (formData: FormData) => Promise<ApiResponse<{ avatar: string }>>
  updatePassword: (data: PasswordUpdateData) => Promise<ApiResponse<PasswordUpdateResponse>>
}

export interface PaginatedResponse<T> {
  items: T[]
  total: number
  page: number
  pageSize: number
  totalPages: number
}

export interface UserSubmission {
  id: number
  problemId: string
  problemTitle: string
  status: string
  language: string
  timeUsed: number
  memoryUsed: number
  submitTime: string
}

export interface UserSolvedProblem {
  id: string
  title: string
  difficulty: string
  solvedAt: string
}

export interface UserContest {
  id: number
  title: string
  startTime: string
  endTime: string
  rank?: number
  score?: number
}
