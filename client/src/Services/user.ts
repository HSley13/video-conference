import { makeRequest } from "./makeRequest";

type GetUserInfoProps = {
  id: string | undefined;
};
export const getUserInfo = async ({ id }: GetUserInfoProps) => {
  const response = await makeRequest({
    url: `/user/userInfo/${id}`,
  });
  return response;
};

type UpdateUserInfoProps = {
  userName: string;
  email: string;
  file?: File;
};
export const updateUserInfo = async ({
  userName,
  email,
  file,
}: UpdateUserInfoProps) => {
  const formData = new FormData();
  formData.append("userName", userName);
  formData.append("email", email);
  if (file) {
    formData.append("image", file);
  }
  const response = await makeRequest({
    url: `/user/updateUserInfo`,
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
    url: `/user/updatePassword`,
    options: {
      method: "PUT",
      data: { newPassword },
    },
  });
  return response;
};

export const deleteUser = async () => {
  const response = await makeRequest({
    url: `/user/deleteUser`,
    options: {
      method: "DELETE",
    },
  });
  return response;
};

export const passwordForgotten = async ({ email }: { email: string }) => {
  const response = await makeRequest({
    url: `/user/passwordForgotten`,
    options: {
      method: "POST",
      data: { email },
    },
  });
  return response;
};
