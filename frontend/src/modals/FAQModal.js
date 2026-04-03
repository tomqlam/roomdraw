import { useState } from "react";

const AGREEMENTS = [
    "Digital Draw is a planning tool meant to help students think through potential Room Draw scenarios. It is not binding, and all plans are subject to change.",
    "Final room assignments are determined during Real Draw and may differ from what is reflected in Digital Draw.",
    "You agree to use Digital Draw in good faith and alignment with the Honor Code.",
    "Students who do not participate in Digital Draw by their assigned deadline will be assigned the worst number of their year (participation is based on the Honor Code).",
    "If you plan to bring a caged animal (e.g., hamster, lizard, etc.) or have an approved ESA (e.g., dog or cat), please note this in the notes section when selecting your suite.",
];

function FAQModal({ isOpen, onClose }) {
    const [checked, setChecked] = useState(AGREEMENTS.map(() => false));

    if (!isOpen) return null;

    const allChecked = checked.every(Boolean);

    const toggle = (i) => {
        const next = [...checked];
        next[i] = !next[i];
        setChecked(next);
    };

    const handleAccept = () => {
        localStorage.setItem("hideWelcomeFAQ", "true");
        onClose();
    };

    return (
        <div className="modal is-active">
            <div className="modal-background"></div>
            <div className="modal-card" style={{ maxWidth: "600px" }}>
                <header
                    className="modal-card-head"
                    style={{
                        backgroundColor: "var(--card-bg)",
                        borderBottom: "1px solid var(--border-color)",
                    }}
                >
                    <p
                        className="modal-card-title"
                        style={{
                            color: "var(--text-color)",
                            fontSize: "1.5rem",
                            fontWeight: "600",
                        }}
                    >
                        Welcome to Digidraw!
                    </p>
                </header>
                <section
                    className="modal-card-body"
                    style={{
                        backgroundColor: "var(--card-bg)",
                        color: "var(--text-color)",
                        padding: "1.5rem",
                    }}
                >
                    <div
                        style={{
                            marginBottom: "1.5rem",
                            padding: "1rem",
                            backgroundColor: "var(--hover-bg, rgba(0,0,0,0.04))",
                            borderRadius: "6px",
                            fontSize: "0.92rem",
                            lineHeight: "1.6",
                        }}
                    >
                        <ol style={{ margin: 0, paddingLeft: "1.25rem" }}>
                            <li style={{ marginBottom: "0.6rem" }}>
                                You are able to pull anyone into any room, not just yourself.
                            </li>
                            <li style={{ marginBottom: "0.6rem" }}>
                                You can only pull into a room if your selected occupants had higher priority than the
                                current occupants, or you clear the room first.
                            </li>
                            <li style={{ marginBottom: "0.6rem" }}>
                                Excessive clearing of rooms will result in a temporary ban. This is to prevent users
                                from evading the pull priority system.
                            </li>
                            <li style={{ marginBottom: "0.6rem" }}>
                                All activity is logged including images uploaded, and any abuse of the system will be
                                investigated and reported to RALs and DSA.
                            </li>
                            <li>
                                If you have any issues, please message the Discord server in the{" "}
                                <strong>#digi-draw</strong> channel.
                            </li>
                        </ol>
                    </div>
                    <p style={{ marginBottom: "1.25rem" }}>
                        Please read through the following agreements and check off each box as you read.
                    </p>
                    <div className="content">
                        {AGREEMENTS.map((text, i) => (
                            <label
                                key={i}
                                style={{
                                    display: "flex",
                                    alignItems: "flex-start",
                                    gap: "0.75rem",
                                    marginBottom: "1.25rem",
                                    cursor: "pointer",
                                    color: "var(--text-color)",
                                }}
                            >
                                <input
                                    type="checkbox"
                                    checked={checked[i]}
                                    onChange={() => toggle(i)}
                                    style={{ marginTop: "0.2rem", flexShrink: 0 }}
                                />
                                <span>{text}</span>
                            </label>
                        ))}
                    </div>
                </section>
                <footer
                    className="modal-card-foot"
                    style={{
                        backgroundColor: "var(--card-bg)",
                        borderTop: "1px solid var(--border-color)",
                        padding: "1rem 1.5rem",
                        flexDirection: "column",
                        alignItems: "flex-start",
                        gap: "0.75rem",
                    }}
                >
                    <p style={{ color: "var(--text-color)", fontSize: "0.9rem", margin: 0 }}>
                        By clicking "I accept", you are confirming that you have read and agree to abide by
                        the instructions listed above. If you have any questions, please email{" "}
                        <a href="mailto:Ral-l@g.hmc.edu">Ral-l@g.hmc.edu</a>.
                    </p>
                    <button
                        className="button is-primary"
                        onClick={handleAccept}
                        disabled={!allChecked}
                        style={{ minWidth: "100px", height: "36px" }}
                    >
                        I accept
                    </button>
                </footer>
            </div>
        </div>
    );
}

export default FAQModal;
