const payload = {
    blocks: [
        {
            type: "section",
            text: {
                type: "plain_text",
                text: "This message was generated with Javascript",
            },
        },
    ],
};
console.log(JSON.stringify(payload)); // Outputs JSON