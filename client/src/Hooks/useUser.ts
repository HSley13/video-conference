type User = {
  id: string;
};
export const useUser = (): User | null => {
  const match = document.cookie.match(/videoConferenceUserId=(?<id>[^;]+);?$/);
  if (match?.groups?.id) {
    return { id: match.groups.id };
  }
  return null;
};

type AccessToken = {
  accessToken: string;
};
export const useAccessToken = (): AccessToken | null => {
  const match = document.cookie.match(/accessToken=(?<accessToken>[^;]+);?$/);
  if (match?.groups?.accessToken) {
    return { accessToken: match.groups.accessToken };
  }
  return null;
};
