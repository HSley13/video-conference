import { makeRequest } from "./makeRequest";

type GetUserInfoProps = {
  id: string | undefined;
};
export const getUserInfo = async ({ id }: GetUserInfoProps) => {
  const response = await makeRequest({
    url: `/auth/userInfo/${id}`,
  });
  return response;
};

type UpdateUserInfoProps = {
  firstName: string;
  lastName: string;
  email: string;
  file?: File;
};
export const updateUserInfo = async ({
  firstName,
  lastName,
  email,
  file,
}: UpdateUserInfoProps) => {
  const formData = new FormData();
  formData.append("firstName", firstName);
  formData.append("lastName", lastName);
  formData.append("email", email);
  if (file) {
    formData.append("image", file);
  }
  const response = await makeRequest({
    url: `/auth/updateUserInfo`,
    options: {
      method: "PUT",
      data: formData,
    },
  });
  return response;
};

type UpdatePasswordProps = {
  newPassword: string;
};
export const updatePassword = async ({ newPassword }: UpdatePasswordProps) => {
  const response = await makeRequest({
    url: `/auth/updatePassword`,
    options: {
      method: "PUT",
      data: { newPassword },
    },
  });
  return response;
};

export const deleteUser = async () => {
  const response = await makeRequest({
    url: `/auth/deleteUser`,
    options: {
      method: "DELETE",
    },
  });
  return response;
};

export const passwordForgotten = async ({ email }: { email: string }) => {
  const response = await makeRequest({
    url: `/auth/passwordForgotten`,
    options: {
      method: "POST",
      data: { email },
    },
  });
  return response;
};
