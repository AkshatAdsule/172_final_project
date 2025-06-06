import * as React from "react";
import { Slot } from "@radix-ui/react-slot";
import styles from "./button.module.css";

export type ButtonVariant =
	| "default"
	| "destructive"
	| "outline"
	| "secondary"
	| "ghost"
	| "link";

export type ButtonSize = "default" | "sm" | "lg" | "icon";

export interface ButtonProps
	extends React.ButtonHTMLAttributes<HTMLButtonElement> {
	variant?: ButtonVariant;
	size?: ButtonSize;
	asChild?: boolean;
}

function Button({
	className,
	variant = "default",
	size = "default",
	asChild = false,
	...props
}: ButtonProps) {
	const Comp = asChild ? Slot : "button";

	// Combine class names
	const buttonClassName = [
		styles.button,
		styles[variant],
		styles[size],
		className,
	]
		.filter(Boolean)
		.join(" ");

	return <Comp data-slot="button" className={buttonClassName} {...props} />;
}

export { Button };
