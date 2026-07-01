"use client";

import { useEffect } from "react";
import { useRouter } from "next/navigation";

export default function GlobalError({
	error,
	reset,
}: {
	error: Error & { digest?: string };
	reset: () => void;
}) {
	const router = useRouter();

	useEffect(() => {
		router.replace("/home");
	}, [router]);

	return (
		<html>
			<body></body>
		</html>
	);
}
