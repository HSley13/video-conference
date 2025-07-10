import React, { useEffect, useState } from "react";

const Card = ({
  visible,
  heading,
  children,
}: {
  visible: boolean;
  heading: string;
  children: React.ReactNode;
}) => (
  <div
    className={`w-full max-w-md transform rounded-3xl bg-white p-6 shadow-lg transition-all duration-700 ease-in-out
      ${visible ? "scale-100 opacity-100 delay-100" : "scale-0 opacity-0"}`}
  >
    <h3 className="mb-4 text-2xl font-semibold">{heading}</h3>
    <div className="space-y-4">{children}</div>
  </div>
);

const Input = ({
  icon,
  ...rest
}: { icon: string } & React.InputHTMLAttributes<HTMLInputElement>) => (
  <label className="relative block">
    <i
      className={`bx ${icon} absolute left-3 top-1/2 -translate-y-1/2 text-xl text-gray-400`}
    />
    <input
      {...rest}
      className="w-full rounded-md bg-gray-100 py-3 pl-10 pr-3 outline-none ring-0 transition focus:ring-2 focus:ring-emerald-600"
    />
  </label>
);

const SubmitButton = ({ children }: { children: React.ReactNode }) => (
  <button className="w-full rounded-lg bg-emerald-600 py-2 text-white transition hover:bg-emerald-700">
    {children}
  </button>
);

const SmallText = ({ children }: { children: React.ReactNode }) => (
  <p className="text-center text-xs">{children}</p>
);

export default function AuthToggle() {
  type Mode = "sign-in" | "sign-up";
  const [mode, setMode] = useState<Mode>("sign-up");

  useEffect(() => {
    const t = setTimeout(() => setMode("sign-in"), 200);
    return () => clearTimeout(t);
  }, []);

  const isSignIn = mode === "sign-in";
  const toggle = () => setMode(isSignIn ? "sign-up" : "sign-in");

  return (
    <div className="relative h-screen overflow-hidden font-[Poppins] text-sm md:text-base">
      <div
        className={`pointer-events-none absolute top-1/2 z-0 h-[200vh] w-[200vh] -translate-y-1/2 rounded-full bg-gradient-to-br from-emerald-600 to-emerald-400 shadow-xl transition-transform duration-700 ease-in-out
          ${isSignIn ? "-translate-x-1/2" : "translate-x-1/2"}`}
      />

      <div className="relative z-10 flex h-full flex-col md:flex-row">
        <div className="flex w-full items-start justify-start p-4 md:w-1/2">
          <Card visible={!isSignIn} heading="Create account">
            <Input icon="bxs-user" placeholder="Username" />
            <Input icon="bx-mail-send" type="email" placeholder="Email" />
            <Input icon="bxs-lock-alt" type="password" placeholder="Password" />
            <Input
              icon="bxs-lock-alt"
              type="password"
              placeholder="Confirm password"
            />
            <SubmitButton>Sign up</SubmitButton>
            <SmallText>
              Already have an account?{" "}
              <b
                onClick={toggle}
                className="cursor-pointer text-emerald-600 hover:underline"
              >
                Sign in here
              </b>
            </SmallText>
          </Card>
        </div>

        <div className="flex w-full flex-grow items-end justify-end p-4 md:w-1/2">
          <Card visible={isSignIn} heading="Welcome back">
            <Input icon="bxs-user" placeholder="Username" />
            <Input icon="bxs-lock-alt" type="password" placeholder="Password" />
            <SubmitButton>Sign in</SubmitButton>
            <p className="text-center text-xs font-semibold">
              Forgot password?
            </p>
            <SmallText>
              Don’t have an account?{" "}
              <b
                onClick={toggle}
                className="cursor-pointer text-emerald-600 hover:underline"
              >
                Sign up here
              </b>
            </SmallText>
          </Card>
        </div>
      </div>

      <div className="pointer-events-none absolute inset-0 z-20 hidden select-none md:flex">
        <div className="flex w-1/2 items-center justify-center">
          <h2
            className={`text-5xl font-extrabold text-white transition-transform duration-700 ease-in-out
              ${isSignIn ? "translate-x-0 opacity-100" : "-translate-x-full opacity-0"}`}
          >
            Welcome
          </h2>
        </div>
        <div className="flex w-1/2 items-center justify-center">
          <h2
            className={`text-5xl font-extrabold text-white transition-transform duration-700 ease-in-out
              ${isSignIn ? "translate-x-full opacity-0" : "translate-x-0 opacity-100"}`}
          >
            Join with us
          </h2>
        </div>
      </div>
    </div>
  );
}
