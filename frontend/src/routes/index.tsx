import { Link, createFileRoute } from "@tanstack/react-router";

export const Route = createFileRoute("/")({
	component: Index,
});

function Index() {
	return (
		<div
			style={{
				display: "flex",
				flexDirection: "column",
				alignItems: "center",
				justifyContent: "center",
				height: "100vh",
				textAlign: "center",
				padding: "20px",
			}}
		>
			<h1 style={{ fontSize: "4rem", margin: "0 0 16px 0", fontWeight: "700" }}>
				B<sup style={{ fontSize: "2rem" }}>3</sup>
			</h1>
			<p
				style={{
					fontSize: "1.2rem",
					color: "#666",
					margin: "0 0 32px 0",
					maxWidth: "600px",
				}}
			>
				Track your rides with real-time GPS monitoring and detailed route
				history. View live location updates and explore your riding patterns.
			</p>
			<div
				style={{
					display: "flex",
					gap: "16px",
					flexWrap: "wrap",
					justifyContent: "center",
				}}
			>
				<Link
					to="/rides/live"
					style={{
						padding: "12px 24px",
						backgroundColor: "#007bff",
						color: "white",
						textDecoration: "none",
						borderRadius: "8px",
						fontWeight: "500",
						transition: "background-color 0.2s",
					}}
				>
					Start Live Tracking
				</Link>
				<Link
					to="/rides"
					style={{
						padding: "12px 24px",
						backgroundColor: "transparent",
						color: "#007bff",
						textDecoration: "none",
						borderRadius: "8px",
						border: "2px solid #007bff",
						fontWeight: "500",
						transition: "all 0.2s",
					}}
				>
					View Ride History
				</Link>
			</div>
		</div>
	);
}
