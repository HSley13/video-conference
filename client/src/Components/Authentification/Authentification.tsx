import { useState, useEffect } from "react";
import { useNavigate } from "react-router-dom";
import { login, register } from "../../Services/auth";
import { useAsyncFn } from "../../Hooks/useAsync";

type AuthentificationState = {
  loginEmail: string;
  lginPassword: string;
  registerUsername: string;
  registerEmail: string;
  registerPassword: string;
  registerConfirmPassword: string;
};

export const Authentification = () => {
  type Mode = "login" | "register";
  const [mode, setMode] = useState<Mode>("login");
  const [errors, setErrors] = useState<Record<string, boolean>>({});
  const [authState, setAuthState] = useState<AuthentificationState>({
    loginEmail: "",
    lginPassword: "",
    registerUsername: "",
    registerEmail: "",
    registerPassword: "",
    registerConfirmPassword: "",
  });

  const loginFn = useAsyncFn(login);
  const registerFn = useAsyncFn(register);
  const navigate = useNavigate();

  const isLogin = mode === "login";
  const toggleMode = () => setMode(isLogin ? "register" : "login");

  useEffect(() => {
    const t = setTimeout(() => setMode("login"), 5000);
    return () => clearTimeout(t);
  }, []);

  const handleInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const { name, value } = e.target;
    setAuthState((prev) => ({ ...prev, [name]: value }));
  };

  const validateFields = (fields: Record<string, string>) => {
    const newErrors: Record<string, boolean> = {};
    Object.keys(fields).forEach((k) => (newErrors[k] = !fields[k]));
    setErrors(newErrors);
    return !Object.values(newErrors).some(Boolean);
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    const fields = isLogin
      ? { email: authState.loginEmail, password: authState.lginPassword }
      : {
          username: authState.registerUsername,
          email: authState.registerEmail,
          password: authState.registerPassword,
          confirmPassword: authState.registerConfirmPassword,
        };

    if (!validateFields(fields)) return;

    try {
      if (isLogin) {
        const res = await loginFn.execute({
          email: fields.email,
          password: fields.password,
        });
        if (res?.status === 200) {
          resetState();
          navigate("/mainwindow");
          return;
        }
      } else {
        const res = await registerFn.execute({
          username: fields.username || "",
          email: fields.email,
          password: fields.password,
        });
        if (res?.status === 200 || res?.status === 201) {
          resetState();
          setMode("login");
          return;
        }
      }
      window.alert("Invalid credentials");
    } catch {
      window.alert("Network or server error");
    }
  };

  const resetState = () =>
    setAuthState({
      loginEmail: "",
      lginPassword: "",
      registerUsername: "",
      registerEmail: "",
      registerPassword: "",
      registerConfirmPassword: "",
    });

  return (
    <div className="relative h-screen overflow-hidden font-[Poppins] text-sm md:text-base">
      <div
        className={`pointer-events-none fixed z-0 h-[200vh] w-[200vh] rounded-full
                    bg-gradient-to-br from-emerald-600 to-emerald-400 shadow-xl
                    transition-transform duration-700 ease-in-out
                    ${
                      isLogin
                        ? "left-0 top-0 -translate-x-1/2 -translate-y-1/2"
                        : "right-0 bottom-0 translate-x-1/2 translate-y-1/2"
                    }`}
      />

      <form
        onSubmit={handleSubmit}
        className="relative z-10 flex h-full flex-col md:flex-row"
      >
        <div className="flex w-full items-start justify-start p-4 md:w-1/2">
          <Card visible={!isLogin} heading="Create an account">
            <Input
              icon="bxs-user"
              placeholder="Username"
              name="registerUsername"
              value={authState.registerUsername}
              onChange={handleInputChange}
              className={`my-3 ${errors.registerUsername ? "invalid" : ""}`}
            />
            <Input
              icon="bx-mail-send"
              type="email"
              placeholder="Email"
              name="registerEmail"
              value={authState.registerEmail}
              onChange={handleInputChange}
              className={`my-3 ${errors.registerEmail ? "invalid" : ""}`}
            />
            <Input
              icon="bxs-lock-alt"
              type="password"
              placeholder="Password"
              name="registerPassword"
              value={authState.registerPassword}
              onChange={handleInputChange}
              className={`my-3 ${errors.registerPassword ? "invalid" : ""}`}
            />
            <Input
              icon="bxs-lock-alt"
              type="password"
              placeholder="Confirm password"
              name="registerConfirmPassword"
              value={authState.registerConfirmPassword}
              onChange={handleInputChange}
              className={`my-3 ${
                errors.registerConfirmPassword ? "invalid" : ""
              }`}
            />

            <SubmitButton>Create account</SubmitButton>
            <SmallText>
              Already have an account?
              <b
                onClick={toggleMode}
                className="cursor-pointer pl-1 text-emerald-600 hover:underline"
              >
                Login
              </b>
            </SmallText>
          </Card>
        </div>

        <div className="flex w-full flex-grow items-end justify-end p-4 md:w-1/2">
          <Card visible={isLogin} heading="Login to your account">
            <Input
              icon="bx-mail-send"
              type="email"
              placeholder="Email"
              name="loginEmail"
              value={authState.loginEmail}
              onChange={handleInputChange}
              className={`my-3 ${errors.loginEmail ? "invalid" : ""}`}
            />
            <Input
              icon="bxs-lock-alt"
              type="password"
              placeholder="Password"
              name="lginPassword"
              value={authState.lginPassword}
              onChange={handleInputChange}
              className={`my-3 ${errors.lginPassword ? "invalid" : ""}`}
            />

            <SubmitButton>Login</SubmitButton>
            <p
              onClick={() => {}}
              className="cursor-pointer text-emerald-600 mt-2 text-center text-sm hover:underline"
            >
              Password Forgotten
            </p>
            <SmallText>
              Don’t have an account?
              <b
                onClick={toggleMode}
                className="cursor-pointer pl-1 text-emerald-600 hover:underline"
              >
                Register
              </b>
            </SmallText>
          </Card>
        </div>

        <div className="pointer-events-none absolute inset-0 z-20 hidden select-none md:flex">
          <div className="flex w-1/2 items-center justify-center">
            <h2
              className={`text-5xl font-extrabold text-white transition-transform duration-700 ease-in-out
                          ${isLogin ? "translate-x-0 opacity-100" : "-translate-x-full opacity-0"}`}
            >
              Welcome Back
            </h2>
          </div>
          <div className="flex w-1/2 items-center justify-center">
            <h2
              className={`text-5xl font-extrabold text-white transition-transform duration-700 ease-in-out
                          ${isLogin ? "translate-x-full opacity-0" : "translate-x-0 opacity-100"}`}
            >
              Join with us
            </h2>
          </div>
        </div>
      </form>
    </div>
  );
};

const Card: React.FC<{
  visible: boolean;
  heading: string;
  children: React.ReactNode;
}> = ({ visible, heading, children }) => (
  <div
    className={`w-full max-w-md transform rounded-3xl bg-white p-6 shadow-lg
                transition-all duration-700 ease-in-out
                ${visible ? "scale-100 opacity-100 delay-100" : "scale-0 opacity-0"}`}
  >
    <h3 className="mb-4 text-2xl font-semibold">{heading}</h3>
    <div className="space-y-4">{children}</div>
  </div>
);

interface InputProps extends React.InputHTMLAttributes<HTMLInputElement> {
  icon: string;
}
const Input = ({ icon, ...rest }: InputProps) => (
  <label className="relative block w-full">
    <i
      className={`bx ${icon} absolute left-3 top-1/2 -translate-y-1/2 text-xl text-gray-400`}
    />
    <input
      {...rest}
      className="w-full rounded-full bg-gray-100 py-3 pl-10 pr-3 outline-none ring-0 transition
                 focus:ring-2 focus:ring-emerald-600"
    />
  </label>
);

const SubmitButton: React.FC<React.ButtonHTMLAttributes<HTMLButtonElement>> = ({
  children,
  ...rest
}) => (
  <button
    {...rest}
    className="w-full !rounded-full bg-emerald-600 py-3 font-semibold text-white transition hover:bg-emerald-700"
  >
    {children}
  </button>
);

const SmallText: React.FC<{ children: React.ReactNode }> = ({ children }) => (
  <p className="text-center text-sm mt-2 text-gray-500">{children}</p>
);
