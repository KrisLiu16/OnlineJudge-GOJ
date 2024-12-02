import request from './config'
import type {
  UserApi,
  LoginData,
  RegisterData,
  User,
  LoginResponse,
  ApiResponse,
  PasswordUpdateData,
  PasswordUpdateResponse,
} from '@/types/user'

export const userApi: UserApi = {
  login: (data: LoginData): Promise<ApiResponse<LoginResponse>> =>
    request.post('/auth/login', data),

  register: (data: RegisterData): Promise<ApiResponse<null>> =>
    request.post('/auth/register', data),

  getProfile: (): Promise<ApiResponse<User>> => request.get('/user/profile'),

  updateProfile: (data: Partial<User>): Promise<ApiResponse<User>> =>
    request.put('/user/profile', data),

  updateAvatar: (formData: FormData): Promise<ApiResponse<{ avatar: string }>> =>
    request.post('/user/avatar', formData),

  updatePassword: (data: PasswordUpdateData): Promise<ApiResponse<PasswordUpdateResponse>> =>
    request.put('/user/password', data),
}

export async function getUserSubmissions(
  username: string,
  page: number = 1,
  pageSize: number = 20,
) {
  const response = await fetch(
    `/api/users/${username}/submissions?page=${page}&pageSize=${pageSize}`,
  )
  if (!response.ok) {
    throw new Error('Failed to fetch submissions')
  }
  return response.json()
}

export async function getUserSolvedProblems(
  username: string,
  page: number = 1,
  pageSize: number = 20,
) {
  const response = await fetch(
    `/api/users/${username}/solved-problems?page=${page}&pageSize=${pageSize}`,
  )
  if (!response.ok) {
    throw new Error('Failed to fetch solved problems')
  }
  return response.json()
}

export async function getUserContests(username: string, page: number = 1, pageSize: number = 20) {
  const response = await fetch(`/api/users/${username}/contests?page=${page}&pageSize=${pageSize}`)
  if (!response.ok) {
    throw new Error('Failed to fetch contests')
  }
  return response.json()
}
