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
		// Redirect to /home on any unhandled error
		router.replace("/home");
	}, [router]);

	return null;
}
