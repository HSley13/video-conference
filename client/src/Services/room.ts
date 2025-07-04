import { makeRequest } from "./makeRequest";

type getRoomProps = {
  id: string;
};
export const getRoom = async ({ id }: getRoomProps) => {
  const response = await makeRequest({
    url: `/room/${id}`,
  });
  return response;
};

type createRoomProps = {
  title: string;
  description?: string;
  password?: string;
};
export const createRoom = async ({
  title,
  description,
  password,
}: createRoomProps) => {
  const response = await makeRequest({
    url: `/room`,
    options: {
      method: "POST",
      data: { title, description, password },
    },
  });
  return response;
};

type joinRoomProps = {
  id: string;
  password?: string;
};
export const joinRoom = async ({ id, password }: joinRoomProps) => {
  const response = await makeRequest({
    url: `/room/${id}`,
    options: {
      method: "PUT",
      data: { password },
    },
  });
  return response;
};

type updateRoomProps = {
  id: string;
  title?: string;
  description?: string;
  password?: string;
  members?: string[];
};
export const updateRoom = async ({
  id,
  title,
  description,
  password,
  members,
}: updateRoomProps) => {
  const response = await makeRequest({
    url: `/room/${id}`,
    options: {
      method: "PATCH",
      data: { title, description, password, members },
    },
  });
  return response;
};

export const deleteRoom = async ({ id }: getRoomProps) => {
  const response = await makeRequest({
    url: `/room/${id}`,
    options: {
      method: "DELETE",
    },
  });
  return response;
};
