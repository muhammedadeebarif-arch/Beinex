export interface AdminProfile {
  id: number;
  name: string;
  email: string;
  createdAt?: string;
}

export interface LoginPayload {
  email: string;
  password: string;
}

export interface LoginResponse {
  token: string;
  admin: AdminProfile;
}
