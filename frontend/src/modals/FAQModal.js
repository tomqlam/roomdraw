import React from 'react';

function FAQModal({ isOpen, onClose })
{
    if (!isOpen) return null;

    return (
        <div className="modal is-active">
            <div className="modal-background" onClick={onClose}></div>
            <div className="modal-card" style={{ maxWidth: '600px' }}>
                <header className="modal-card-head" style={{
                    backgroundColor: 'var(--card-bg)',
                    borderBottom: '1px solid var(--border-color)'
                }}>
                    <p className="modal-card-title" style={{
                        color: 'var(--text-color)',
                        fontSize: '1.5rem',
                        fontWeight: '600'
                    }}>Welcome to DigiDraw!</p>
                    <button className="delete" aria-label="close" onClick={onClose}></button>
                </header>
                <section className="modal-card-body" style={{
                    backgroundColor: 'var(--card-bg)',
                    color: 'var(--text-color)',
                    padding: '1.5rem'
                }}>
                    <div className="content">
                        <ol style={{
                            listStyleType: 'none',
                            padding: 0,
                            margin: 0,
                            counterReset: 'item'
                        }}>
                            <li style={{
                                marginBottom: '1.25rem',
                                paddingLeft: '2rem',
                                position: 'relative',
                                counterIncrement: 'item'
                            }}>
                                <span style={{
                                    position: 'absolute',
                                    left: 0,
                                    fontWeight: '600'
                                }}>{1}.</span>
                                You are able to pull anyone into any room, not just yourself.
                            </li>

                            <li style={{
                                marginBottom: '1.25rem',
                                paddingLeft: '2rem',
                                position: 'relative',
                                counterIncrement: 'item'
                            }}>
                                <span style={{
                                    position: 'absolute',
                                    left: 0,
                                    fontWeight: '600'
                                }}>{2}.</span>
                                You can only pull into a room if:
                                <ul style={{
                                    marginTop: '0.75rem',
                                    marginLeft: '1rem',
                                    listStyleType: 'disc'
                                }}>
                                    <li style={{ marginBottom: '0.5rem' }}>your selected occupants had higher priority than the current occupants</li>
                                    <li>or, you clear the room first.</li>
                                </ul>
                            </li>

                            <li style={{
                                marginBottom: '1.25rem',
                                paddingLeft: '2rem',
                                position: 'relative',
                                counterIncrement: 'item'
                            }}>
                                <span style={{
                                    position: 'absolute',
                                    left: 0,
                                    fontWeight: '600'
                                }}>{3}.</span>
                                Excessive clearing of rooms will result in a temporary ban. This is to prevent users from evading the pull priority system.
                            </li>

                            <li style={{
                                marginBottom: '1.25rem',
                                paddingLeft: '2rem',
                                position: 'relative',
                                counterIncrement: 'item'
                            }}>
                                <span style={{
                                    position: 'absolute',
                                    left: 0,
                                    fontWeight: '600'
                                }}>{4}.</span>
                                All activity is logged including images uploaded and any abuse of the system will be investigated and reported to RALs and DSA.
                            </li>

                            <li style={{
                                marginBottom: '0',
                                paddingLeft: '2rem',
                                position: 'relative',
                                counterIncrement: 'item'
                            }}>
                                <span style={{
                                    position: 'absolute',
                                    left: 0,
                                    fontWeight: '600'
                                }}>{5}.</span>
                                If you have any issues, please message the discord server in the #digi-draw channel.
                            </li>
                        </ol>
                    </div>
                </section>
                <footer className="modal-card-foot" style={{
                    backgroundColor: 'var(--card-bg)',
                    borderTop: '1px solid var(--border-color)',
                    padding: '1rem 1.5rem',
                    display: 'flex',
                    alignItems: 'center',
                    gap: '1rem'
                }}>
                    <button
                        className="button is-primary"
                        onClick={onClose}
                        style={{
                            minWidth: '100px',
                            height: '36px',
                            display: 'flex',
                            alignItems: 'center',
                            justifyContent: 'center'
                        }}
                    >
                        Got it!
                    </button>
                    <label className="checkbox" style={{
                        display: 'flex',
                        alignItems: 'center',
                        gap: '0.5rem',
                        color: 'var(--text-color)'
                    }}>
                        <input
                            type="checkbox"
                            onChange={(e) =>
                            {
                                if (e.target.checked)
                                {
                                    localStorage.setItem('hideWelcomeFAQ', 'true');
                                }
                            }}
                            style={{
                                margin: 0
                            }}
                        />
                        Don't show this again
                    </label>
                </footer>
            </div>
        </div>
    );
}

export default FAQModal; 