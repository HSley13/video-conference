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
  username: string;
  email: string;
  password: string;
};
export const register = async ({
  username,
  email,
  password,
}: RegisterProps) => {
  const response = await makeRequest({
    url: "/auth/register",
    options: {
      method: "POST",
      data: { username, email, password },
    },
  });
  return response;
};
