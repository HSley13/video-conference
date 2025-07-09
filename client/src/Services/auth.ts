import { makeRequest } from "./makeRequest";

type SignInProps = {
  email: string;
  password: string;
};
export const signIn = async ({ email, password }: SignInProps) => {
  const response = await makeRequest({
    url: "/auth/signIn",
    options: {
      method: "POST",
      data: { email, password },
    },
  });
  return response;
};

type SignUpProps = {
  firstName: string;
  lastName: string;
  email: string;
  password: string;
};
export const signUp = async ({
  firstName,
  lastName,
  email,
  password,
}: SignUpProps) => {
  const response = await makeRequest({
    url: "/auth/signUp",
    options: {
      method: "POST",
      data: { firstName, lastName, email, password },
    },
  });
  return response;
};
