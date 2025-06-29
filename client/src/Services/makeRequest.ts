import axios, { AxiosRequestConfig, AxiosResponse, AxiosError } from "axios";

type MakeRequestProps = {
  url: string;
  options?: AxiosRequestConfig;
};

const api = axios.create({
  baseURL: import.meta.env.VITE_API_URL,
  withCredentials: true,
});

export const makeRequest = async ({ url, options }: MakeRequestProps) => {
  try {
    const response: AxiosResponse = await api(url, options);
    return response.data;
  } catch (error) {
    const axiosError = error as AxiosError;

    if (axiosError.response) {
      const responseData = axiosError.response.data;

      if (
        typeof responseData === "object" &&
        responseData !== null &&
        "message" in responseData
      ) {
        return Promise.reject(responseData.message);
      } else if (typeof responseData === "string") {
        return Promise.reject(responseData);
      } else {
        return Promise.reject("An error occurred");
      }
    } else if (axiosError.request) {
      return Promise.reject("No response received from the server");
    } else {
      return Promise.reject("An error occurred while setting up the request");
    }
  }
};
