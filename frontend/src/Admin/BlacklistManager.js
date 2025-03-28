import React, { useEffect, useState } from 'react';

function BlacklistManager()
{
    const [blacklistedUsers, setBlacklistedUsers] = useState([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState(null);
    const [removing, setRemoving] = useState({});

    useEffect(() =>
    {
        fetchBlacklistedUsers();
    }, []);

    const fetchBlacklistedUsers = () =>
    {
        setLoading(true);
        fetch(`${process.env.REACT_APP_API_URL}/admin/blacklisted-users`, {
            headers: {
                'Authorization': `Bearer ${localStorage.getItem('jwt')}`
            }
        })
            .then(response =>
            {
                if (!response.ok)
                {
                    throw new Error('Failed to fetch blacklisted users');
                }
                return response.json();
            })
            .then(data =>
            {
                setBlacklistedUsers(data);
                setLoading(false);
            })
            .catch(err =>
            {
                setError(err.message);
                setLoading(false);
            });
    };

    const removeFromBlacklist = (email) =>
    {
        setRemoving(prev => ({ ...prev, [email]: true }));

        fetch(`${process.env.REACT_APP_API_URL}/admin/blacklist/remove/${email}`, {
            method: 'POST',
            headers: {
                'Authorization': `Bearer ${localStorage.getItem('jwt')}`
            }
        })
            .then(response =>
            {
                if (!response.ok)
                {
                    throw new Error('Failed to remove user from blacklist');
                }
                return response.json();
            })
            .then(() =>
            {
                // Remove user from the list
                setBlacklistedUsers(prev => prev.filter(user => user.email !== email));
                setRemoving(prev => ({ ...prev, [email]: false }));
            })
            .catch(err =>
            {
                setError(err.message);
                setRemoving(prev => ({ ...prev, [email]: false }));
            });
    };

    if (loading)
    {
        return (
            <div className="section">
                <div className="container">
                    <h1 className="title has-text-centered">Admin Dashboard</h1>
                    <div className="box has-text-centered" style={{ padding: "3rem" }}>
                        <span className="icon is-large">
                            <i className="fas fa-circle-notch fa-spin fa-2x"></i>
                        </span>
                        <p className="mt-3">Loading blacklisted users...</p>
                    </div>
                </div>
            </div>
        );
    }

    if (error)
    {
        return (
            <div className="section">
                <div className="container">
                    <h1 className="title has-text-centered">Admin Dashboard</h1>
                    <div className="box">
                        <div className="notification is-danger">
                            <p><strong>Error:</strong> {error}</p>
                            <button className="button is-small is-light mt-2" onClick={fetchBlacklistedUsers}>
                                <span className="icon">
                                    <i className="fas fa-sync-alt"></i>
                                </span>
                                <span>Retry</span>
                            </button>
                        </div>
                    </div>
                </div>
            </div>
        );
    }

    return (
        <div className="section">
            <div className="container">
                <div className="box">
                    <div className="mb-4">
                        <div className="level">
                            <div className="level-left">
                                <div className="level-item">
                                    <h3 className="title is-4 mb-0">
                                        <span className="icon mr-4 has-text-primary">
                                            <i className="fas fa-user-lock"></i>
                                        </span>
                                        Blacklisted Users
                                    </h3>
                                </div>
                            </div>
                            <div className="level-right">
                                <div className="level-item">
                                    <button
                                        className="button is-light"
                                        onClick={fetchBlacklistedUsers}
                                        title="Refresh data"
                                    >
                                        <span className="icon">
                                            <i className="fas fa-sync-alt"></i>
                                        </span>
                                        <span>Refresh</span>
                                    </button>
                                </div>
                            </div>
                        </div>
                        <p className="subtitle is-6 mt-2">
                            Users who have been blacklisted for excessive room clearing. These users cannot perform any room actions until removed from the blacklist.
                        </p>
                    </div>

                    {blacklistedUsers.length === 0 ? (
                        <div className="notification is-info is-light">
                            <span className="icon mr-2">
                                <i className="fas fa-info-circle"></i>
                            </span>
                            No blacklisted users found.
                        </div>
                    ) : (
                        <div className="table-container">
                            <table className="table is-fullwidth is-hoverable">
                                <thead>
                                    <tr>
                                        <th>Email</th>
                                        <th>Room Clears</th>
                                        <th>Blacklisted At</th>
                                        <th>Reason</th>
                                        <th>Actions</th>
                                    </tr>
                                </thead>
                                <tbody>
                                    {blacklistedUsers.map(user => (
                                        <tr key={user.email}>
                                            <td>
                                                <strong>{user.email}</strong>
                                            </td>
                                            <td>
                                                <span className="tag is-danger">{user.clearRoomCount}</span>
                                            </td>
                                            <td>{new Date(user.blacklistedAt).toLocaleString()}</td>
                                            <td>{user.reason}</td>
                                            <td>
                                                <button
                                                    className={`button is-primary ${removing[user.email] ? 'is-loading' : ''}`}
                                                    onClick={() => removeFromBlacklist(user.email)}
                                                    disabled={removing[user.email]}
                                                >
                                                    <span className="icon">
                                                        <i className="fas fa-user-check"></i>
                                                    </span>
                                                    <span>Remove</span>
                                                </button>
                                            </td>
                                        </tr>
                                    ))}
                                </tbody>
                            </table>
                        </div>
                    )}
                </div>
            </div>
        </div>
    );
}

export default BlacklistManager; 