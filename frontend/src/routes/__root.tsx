import { Outlet, createRootRoute } from "@tanstack/react-router";
import "./root.css";

export const Route = createRootRoute({
	component: () => (
		<>
			<Outlet />
		</>
	),
});
