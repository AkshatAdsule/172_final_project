import { Outlet, createRootRoute } from "@tanstack/react-router";
import { TanStackRouterDevtools } from "@tanstack/react-router-devtools";
import { LatLngProvider } from "../providers/LatLngProvider";

import "../App.css";

export const Route = createRootRoute({
	component: () => (
		<>
			<LatLngProvider url="ws://localhost:8080">
				<Outlet />
			</LatLngProvider>
			<TanStackRouterDevtools />
		</>
	),
});
