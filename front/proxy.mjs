import express from "express";
import cors from "cors";
import fetch from "node-fetch";

const app = express();
app.use(cors());

app.get("/proxy", async (req, res) => {
    const { search, stars } = req.query;

    const url = `https://jn5d9mrgm3.execute-api.eu-north-1.amazonaws.com/prod/search?search=${encodeURIComponent(search)}&stars=${encodeURIComponent(stars)}`;
    console.log(`Forwarding request to: ${url}`);

    try {
        const response = await fetch(url);
        const data = await response.json();

        // Handle empty responses gracefully
        if (!data || data.length === 0) {
            console.log("No matching results found.");
            return res.status(200).json([]);
        }

        res.status(200).json(data);
    } catch (error) {
        console.error("Error in proxy server:", error.message);
        res.status(500).json({ error: "Failed to fetch data from AWS Lambda" });
    }
});

const PORT = 3000;
app.listen(PORT, () => {
    console.log(`Proxy server running on http://localhost:${PORT}`);
});

