import { createFileRoute, Link } from "@tanstack/react-router";
import "./styles/index.css";

export const Route = createFileRoute("/rides/")({
	component: RouteComponent,
});

function RouteComponent() {
	return (
		<>
			<div className="empty-state">
				<span>
					View <Link to="/rides">live location </Link>
					or select a ride from your history.
				</span>
			</div>
		</>
	);
}
