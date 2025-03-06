import { VariantProps, cva } from "class-variance-authority";
import { ComponentProps } from "react";
import { twMerge } from "tailwind-merge";

export const buttonStyles = cva(["transition-colors"], {
  variants: {
    variant: {
      default: ["bg-secondary", "hover:bg-secondary-hover"],
      dark: ["bg-secondary-dark", "hover:bg-secondary-dark-hover"],
    },
  },
  defaultVariants: { variant: "default" },
});

type ButtonProps = VariantProps<typeof buttonStyles> & ComponentProps<"button">;

export const SideBarButton = ({
  variant,
  className,
  ...props
}: ButtonProps) => {
  return (
    <button
      {...props}
      className={twMerge(buttonStyles({ variant }), className)}
    ></button>
  );
};
