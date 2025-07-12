import { makeRequest } from "./makeRequest";

type LoginProps = {
  email: string;
  password: string;
};
export const login = async ({ email, password }: LoginProps) => {
  const response = await makeRequest({
    url: "/auth/login",
    options: {
      method: "POST",
      data: { email, password },
    },
  });
  return response;
};

type RegisterProps = {
  userName: string;
  email: string;
  password: string;
};
export const register = async ({
  userName,
  email,
  password,
}: RegisterProps) => {
  const response = await makeRequest({
    url: "/auth/register",
    options: {
      method: "POST",
      data: { userName, email, password },
    },
  });
  return response;
};
