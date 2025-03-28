import React, { useContext, useEffect, useState } from 'react';
import Select from 'react-select';
import { MyContext } from '../MyContext';

function SearchPage()
{
    const {
        rooms,
        userMap,
        handleTakeMeThere,
        handleErrorFromTokenExpiry,
        isDarkMode,
        setCurrPage
    } = useContext(MyContext);

    const [searchType, setSearchType] = useState('rooms'); // 'rooms' or 'people'
    const [loading, setLoading] = useState(false);
    const [showLoading, setShowLoading] = useState(false); // Separate state for showing loading animation
    const [dormFilter, setDormFilter] = useState([]);
    const [capacityFilter, setCapacityFilter] = useState([]);
    const [yearFilter, setYearFilter] = useState([]);
    const [inDormFilter, setInDormFilter] = useState([]);
    const [drawNumberFilter, setDrawNumberFilter] = useState({ min: '', max: '' });
    const [results, setResults] = useState([]);
    const [showFilters, setShowFilters] = useState(true);
    const [page, setPage] = useState(1);
    const [itemsPerPage, setItemsPerPage] = useState(10);
    const [sortConfig, setSortConfig] = useState({ key: '', direction: '' });
    const [totalPages, setTotalPages] = useState(1);
    const [totalRecords, setTotalRecords] = useState(0);

    // Derived data
    const [dormOptions, setDormOptions] = useState([]);
    const [capacityOptions, setCapacityOptions] = useState([]);
    const [inDormOptions, setInDormOptions] = useState([]);

    // Custom styles for react-select to support dark mode
    const selectStyles = {
        control: (baseStyles) => ({
            ...baseStyles,
            backgroundColor: isDarkMode ? 'var(--card-bg)' : 'white',
            borderColor: 'var(--border-color)',
            boxShadow: 'var(--input-shadow)',
            '&:hover': {
                borderColor: 'var(--primary-color)',
            },
        }),
        option: (baseStyles, state) => ({
            ...baseStyles,
            backgroundColor: isDarkMode
                ? state.isSelected
                    ? 'var(--primary-color)'
                    : state.isFocused
                        ? 'var(--card-bg-hover)'
                        : 'var(--card-bg)'
                : state.isSelected
                    ? 'var(--primary-color)'
                    : state.isFocused
                        ? '#f5f5f5'
                        : 'white',
            color: isDarkMode
                ? state.isSelected
                    ? 'white'
                    : 'var(--text-color)'
                : state.isSelected
                    ? 'white'
                    : 'inherit',
            cursor: 'pointer',
        }),
        menu: (baseStyles) => ({
            ...baseStyles,
            backgroundColor: isDarkMode ? 'var(--card-bg)' : 'white',
            boxShadow: 'var(--box-shadow)',
            zIndex: 10,
        }),
        input: (baseStyles) => ({
            ...baseStyles,
            color: isDarkMode ? 'var(--text-color)' : 'inherit',
        }),
        singleValue: (baseStyles) => ({
            ...baseStyles,
            color: isDarkMode ? 'var(--text-color)' : 'inherit',
        }),
        placeholder: (baseStyles) => ({
            ...baseStyles,
            color: isDarkMode ? 'var(--text-muted)' : 'hsl(0, 0%, 50%)',
        }),
    };

    // Utility function to convert dorm ID to dorm name
    const getDormNameFromId = (dormId) =>
    {
        if (!dormId || dormId === 0) return "None";

        const dormMapping = {
            1: 'East',
            2: 'North',
            3: 'South',
            4: 'West',
            5: 'Atwood',
            6: 'Sontag',
            7: 'Case',
            8: 'Drinkward',
            9: 'Linde'
        };

        return dormMapping[dormId] || `Unknown (${dormId})`;
    };

    // Handle delayed loading indicator
    useEffect(() =>
    {
        let timer;
        if (loading)
        {
            // Only show loading spinner after 1 second
            timer = setTimeout(() =>
            {
                setShowLoading(true);
            }, 1000);
        } else
        {
            setShowLoading(false);
        }

        // Cleanup timer on unmount or when loading state changes
        return () =>
        {
            if (timer) clearTimeout(timer);
        };
    }, [loading]);

    // Get unique dorms and capacities from room data
    useEffect(() =>
    {
        if (rooms && rooms.length > 0)
        {
            // Extract unique dorm names
            const uniqueDorms = [...new Set(rooms.map(room => room.DormName))];
            setDormOptions(uniqueDorms.map(dorm => ({ value: dorm, label: dorm })));

            // Extract unique capacities
            const uniqueCapacities = [...new Set(rooms.map(room => room.MaxOccupancy))].sort((a, b) => a - b);
            setCapacityOptions(uniqueCapacities.map(cap => ({ value: cap, label: `${cap} person${cap !== 1 ? 's' : ''}` })));
        }

        // Create in-dorm options from rooms
        if (rooms && rooms.length > 0)
        {
            const uniqueDorms = [...new Set(rooms.map(room => room.DormName))];
            setInDormOptions(uniqueDorms.map(dorm => ({ value: dorm, label: dorm })));
        }
    }, [rooms]);

    // Year options corrected - no "freshman" in the database
    const yearOptions = [
        { value: 'sophomore', label: 'Sophomore' },
        { value: 'junior', label: 'Junior' },
        { value: 'senior', label: 'Senior' }
    ];

    useEffect(() =>
    {
        if (results.length > 0)
        {
            // Fetch data when page, itemsPerPage, or sorting changes
            handleSearch();
        }
    }, [page, itemsPerPage, sortConfig]);

    const handleSearch = () =>
    {
        setLoading(true);
        // showLoading will be set to true after 1 second by the useEffect

        if (searchType === 'rooms')
        {
            // Build query parameters for rooms search
            const params = new URLSearchParams();
            params.append('page', page);
            params.append('limit', itemsPerPage);
            params.append('empty_only', 'true'); // Always filter for available rooms

            // Add dorm filter if selected
            if (dormFilter.length > 0)
            {
                params.append('dorm', dormFilter[0].value); // For now, use only the first selected dorm
            }

            // Add capacity filter if selected
            if (capacityFilter.length > 0)
            {
                params.append('capacity', capacityFilter[0].value); // For now, use only the first selected capacity
            }

            // Add sorting parameters
            if (sortConfig.key)
            {
                let serverSortKey = sortConfig.key;
                // Map frontend keys to backend keys
                const keyMapping = {
                    'DormName': 'dorm_name',
                    'RoomID': 'room_id',
                    'capacity': 'max_occupancy',
                    'currentOccupants': 'current_occupancy',
                    'spacesAvailable': 'max_occupancy' // Special case, we'll sort by max_occupancy
                };

                if (keyMapping[sortConfig.key])
                {
                    serverSortKey = keyMapping[sortConfig.key];
                }

                params.append('sort_by', serverSortKey);
                params.append('sort_order', sortConfig.direction);
            }

            // Fetch rooms from server with pagination and sorting
            fetch(`${process.env.REACT_APP_API_URL}/search/rooms?${params.toString()}`, {
                headers: {
                    'Authorization': `Bearer ${localStorage.getItem('jwt')}`
                }
            })
                .then(response => response.json())
                .then(data =>
                {
                    if (handleErrorFromTokenExpiry(data))
                    {
                        setLoading(false);
                        return;
                    }

                    setResults(data.rooms || []);
                    setTotalPages(data.total_pages || 1);
                    setTotalRecords(data.total || 0);
                    setLoading(false);
                })
                .catch(error =>
                {
                    console.error('Error fetching rooms:', error);
                    setLoading(false);
                    setResults([]);
                });
        } else
        {
            // Build query parameters for users search
            const params = new URLSearchParams();
            params.append('page', page);
            params.append('limit', itemsPerPage);

            // Add year filter if selected
            if (yearFilter.length > 0)
            {
                params.append('year', yearFilter[0].value); // For now, use only the first selected year
            }

            // Add draw number range if specified
            if (drawNumberFilter.min)
            {
                params.append('min_draw_number', drawNumberFilter.min);
            }
            if (drawNumberFilter.max)
            {
                params.append('max_draw_number', drawNumberFilter.max);
            }

            // Add in-dorm filter if selected
            if (inDormFilter.length > 0)
            {
                // Get dorm ID from the dorm name
                const dormName = inDormFilter[0].value;
                const dormRoom = rooms.find(room => room.DormName === dormName);
                if (dormRoom)
                {
                    params.append('in_dorm', dormRoom.Dorm);
                }
            }

            // Add sorting parameters
            if (sortConfig.key)
            {
                let serverSortKey = sortConfig.key;
                // Map frontend keys to backend keys
                const keyMapping = {
                    'FirstName': 'first_name',
                    'LastName': 'last_name',
                    'year': 'year',
                    'drawNumber': 'draw_number'
                };

                if (keyMapping[sortConfig.key])
                {
                    serverSortKey = keyMapping[sortConfig.key];
                }

                params.append('sort_by', serverSortKey);
                params.append('sort_order', sortConfig.direction);
            }

            // Fetch users from server with pagination and sorting
            fetch(`${process.env.REACT_APP_API_URL}/search/users?${params.toString()}`, {
                headers: {
                    'Authorization': `Bearer ${localStorage.getItem('jwt')}`
                }
            })
                .then(response => response.json())
                .then(data =>
                {
                    if (handleErrorFromTokenExpiry(data))
                    {
                        setLoading(false);
                        return;
                    }

                    setResults(data.users || []);
                    setTotalPages(data.total_pages || 1);
                    setTotalRecords(data.total || 0);
                    setLoading(false);
                })
                .catch(error =>
                {
                    console.error('Error fetching users:', error);
                    setLoading(false);
                    setResults([]);
                });
        }
    };

    const requestSort = (key) =>
    {
        // Set loading state before changing sort to prevent flicker
        setLoading(true);
        // showLoading will be set to true after 1 second by the useEffect

        let direction = 'ascending';
        if (sortConfig.key === key && sortConfig.direction === 'ascending')
        {
            direction = 'descending';
        }
        setSortConfig({ key, direction });
        setPage(1); // Reset to first page when sorting changes

        // We don't need to call handleSearch here as the useEffect will trigger it
        // This prevents double loading state
    };

    const getClassNamesFor = (name) =>
    {
        if (!sortConfig) return;
        return sortConfig.key === name ? sortConfig.direction : '';
    };

    const resetFilters = () =>
    {
        setDormFilter([]);
        setCapacityFilter([]);
        setYearFilter([]);
        setInDormFilter([]);
        setDrawNumberFilter({ min: '', max: '' });
        setResults([]);
        setSortConfig({ key: '', direction: '' });
        setPage(1);
        setTotalPages(1);
        setTotalRecords(0);
    };

    const toggleFilters = () =>
    {
        setShowFilters(!showFilters);
    };

    const navigateToRoom = (roomInfo) =>
    {
        const locationString = `${roomInfo.DormName} ${roomInfo.RoomID}`;
        // Switch back to Home view before navigating to room
        setCurrPage('Home');
        // Use setTimeout to ensure the home view is loaded before navigating
        setTimeout(() =>
        {
            handleTakeMeThere(locationString, false);
        }, 50);
    };

    // Create pagination controls
    const pagination = () =>
    {
        const pages = [];

        // Reduce excessive ellipsis when there are few pages
        const showEllipsis = totalPages > 7;
        const ellipsis = <li key="ellipsis">
            <span className="pagination-ellipsis">&hellip;</span>
        </li>;

        // Logic to determine range of pages to show
        let startPage = 1;
        let endPage = totalPages;

        if (totalPages > 7)
        {
            if (page <= 4)
            {
                // Near the beginning
                endPage = 5;
            } else if (page >= totalPages - 3)
            {
                // Near the end
                startPage = totalPages - 4;
            } else
            {
                // In the middle
                startPage = page - 2;
                endPage = page + 2;
            }
        }

        // First page with ellipsis if needed
        if (startPage > 1)
        {
            pages.push(
                <li key={1}>
                    <a className="pagination-link" onClick={() => setPage(1)}>1</a>
                </li>
            );

            if (startPage > 2 && showEllipsis)
            {
                pages.push(ellipsis);
            }
        }

        // Page numbers
        for (let i = startPage; i <= endPage; i++)
        {
            pages.push(
                <li key={i}>
                    <a
                        className={`pagination-link ${i === page ? 'is-current' : ''}`}
                        onClick={() => setPage(i)}
                    >
                        {i}
                    </a>
                </li>
            );
        }

        // Last page with ellipsis if needed
        if (endPage < totalPages)
        {
            if (endPage < totalPages - 1 && showEllipsis)
            {
                pages.push(ellipsis);
            }

            pages.push(
                <li key={totalPages}>
                    <a className="pagination-link" onClick={() => setPage(totalPages)}>
                        {totalPages}
                    </a>
                </li>
            );
        }

        return pages;
    };

    return (
        <div className="section">
            <div className="container">
                <h1 className="title has-text-centered mb-5">Search</h1>

                <div className="box mb-4 search-box">
                    <div className="buttons has-addons is-centered mb-3">
                        <button
                            className={`button ${searchType === 'rooms' ? 'is-primary' : isDarkMode ? 'is-dark' : 'is-light'}`}
                            onClick={() => { setSearchType('rooms'); resetFilters(); }}
                            style={{ borderRadius: '8px 0 0 8px', transition: 'all 0.3s ease' }}
                        >
                            <span className="icon">
                                <i className="fas fa-door-open"></i>
                            </span>
                            <span>Find Empty Rooms</span>
                        </button>
                        <button
                            className={`button ${searchType === 'people' ? 'is-primary' : isDarkMode ? 'is-dark' : 'is-light'}`}
                            onClick={() => { setSearchType('people'); resetFilters(); }}
                            style={{ borderRadius: '0 8px 8px 0', transition: 'all 0.3s ease' }}
                        >
                            <span className="icon">
                                <i className="fas fa-user"></i>
                            </span>
                            <span>Find People</span>
                        </button>
                    </div>

                    <div className="is-flex is-justify-content-flex-end mb-3">
                        <button
                            className={`button is-small ${isDarkMode ? 'is-dark' : 'is-light'}`}
                            onClick={toggleFilters}
                            title={showFilters ? "Hide filters" : "Show filters"}
                            style={{ borderRadius: '6px', transition: 'all 0.3s ease' }}
                        >
                            <span className="icon">
                                <i className={`fas ${showFilters ? 'fa-chevron-up' : 'fa-chevron-down'}`}></i>
                            </span>
                            <span>{showFilters ? 'Hide Filters' : 'Show Filters'}</span>
                        </button>
                    </div>

                    {showFilters && (
                        <div className="px-4 py-3 search-filter-panel">
                            {searchType === 'rooms' ? (
                                <div className="columns is-centered">
                                    <div className="column is-5">
                                        <label className="label dark-mode-label">Dorm</label>
                                        <Select
                                            className="react-select"
                                            classNamePrefix="react-select"
                                            placeholder="All Dorms"
                                            options={dormOptions}
                                            value={dormFilter.length > 0 ? dormFilter[0] : null}
                                            onChange={(option) => setDormFilter(option ? [option] : [])}
                                            isClearable
                                            styles={selectStyles}
                                        />
                                    </div>
                                    <div className="column is-5">
                                        <label className="label dark-mode-label">Room Capacity</label>
                                        <Select
                                            className="react-select"
                                            classNamePrefix="react-select"
                                            placeholder="Any Size"
                                            options={capacityOptions}
                                            value={capacityFilter.length > 0 ? capacityFilter[0] : null}
                                            onChange={(option) => setCapacityFilter(option ? [option] : [])}
                                            isClearable
                                            styles={selectStyles}
                                        />
                                    </div>
                                </div>
                            ) : (
                                <div className="columns mb-0">
                                    <div className="column is-one-third">
                                        <label className="label dark-mode-label">Year</label>
                                        <Select
                                            className="react-select"
                                            classNamePrefix="react-select"
                                            placeholder="All Years"
                                            options={yearOptions}
                                            value={yearFilter.length > 0 ? yearFilter[0] : null}
                                            onChange={(option) => setYearFilter(option ? [option] : [])}
                                            isClearable
                                            styles={selectStyles}
                                        />
                                    </div>
                                    <div className="column is-one-third">
                                        <label className="label dark-mode-label">In-Dorm Preference</label>
                                        <Select
                                            className="react-select"
                                            classNamePrefix="react-select"
                                            placeholder="Any Dorm"
                                            options={inDormOptions}
                                            value={inDormFilter.length > 0 ? inDormFilter[0] : null}
                                            onChange={(option) => setInDormFilter(option ? [option] : [])}
                                            isClearable
                                            styles={selectStyles}
                                        />
                                    </div>
                                    <div className="column is-one-third">
                                        <label className="label dark-mode-label">Draw Number Range</label>
                                        <div className="field has-addons">
                                            <div className="control is-expanded">
                                                <input
                                                    className="input search-number-input"
                                                    type="number"
                                                    placeholder="Min"
                                                    min="1"
                                                    value={drawNumberFilter.min}
                                                    onChange={(e) => setDrawNumberFilter({ ...drawNumberFilter, min: e.target.value })}
                                                />
                                            </div>
                                            <div className="control">
                                                <a className="button is-static search-static-button">to</a>
                                            </div>
                                            <div className="control is-expanded">
                                                <input
                                                    className="input search-number-input-right"
                                                    type="number"
                                                    placeholder="Max"
                                                    min="1"
                                                    value={drawNumberFilter.max}
                                                    onChange={(e) => setDrawNumberFilter({ ...drawNumberFilter, max: e.target.value })}
                                                />
                                            </div>
                                        </div>
                                    </div>
                                </div>
                            )}

                            <div className="field is-grouped is-flex is-justify-content-center mt-4">
                                <div className="control">
                                    <button
                                        className={`button is-primary ${loading ? 'is-loading' : ''}`}
                                        onClick={() =>
                                        {
                                            setPage(1); // Reset to first page when searching
                                            handleSearch();
                                        }}
                                        style={{ borderRadius: '8px', transition: 'all 0.3s ease', width: '120px' }}
                                    >
                                        <span className="icon">
                                            <i className="fas fa-search"></i>
                                        </span>
                                        <span>Search</span>
                                    </button>
                                </div>
                                <div className="control">
                                    <button
                                        className={`button ${isDarkMode ? 'is-dark' : 'is-light'}`}
                                        onClick={resetFilters}
                                        style={{ borderRadius: '8px', transition: 'all 0.3s ease' }}
                                    >
                                        <span className="icon">
                                            <i className="fas fa-times"></i>
                                        </span>
                                        <span>Reset</span>
                                    </button>
                                </div>
                            </div>
                        </div>
                    )}
                </div>

                {/* Results section */}
                <div className="box search-results-box">
                    <div className="level">
                        <div className="level-left">
                            <div className="level-item">
                                <h3 className="title is-4 mb-0">
                                    <span className="icon mr-4 has-text-primary">
                                        <i className={`fas ${searchType === 'rooms' ? 'fa-door-open' : 'fa-user'}`}></i>
                                    </span>
                                    {searchType === 'rooms' ? 'Available Rooms' : 'People'}
                                </h3>
                            </div>
                        </div>
                        <div className="level-right">
                            <div className="level-item">
                                <div className="is-flex is-align-items-center">
                                    <div className="select is-small mr-2">
                                        <select
                                            value={itemsPerPage}
                                            onChange={(e) =>
                                            {
                                                setItemsPerPage(Number(e.target.value));
                                                setPage(1); // Reset to first page when changing items per page
                                            }}
                                            className="dark-mode-select"
                                        >
                                            <option value={5}>5 per page</option>
                                            <option value={10}>10 per page</option>
                                            <option value={25}>25 per page</option>
                                            <option value={50}>50 per page</option>
                                        </select>
                                    </div>
                                    <span className="results-count">{totalRecords} results</span>
                                </div>
                            </div>
                        </div>
                    </div>

                    <div className="results-content" style={{ minHeight: '300px', position: 'relative' }}>
                        {loading && showLoading ? (
                            <div className="has-text-centered search-loading-overlay">
                                <span className="icon is-large">
                                    <i className="fas fa-circle-notch fa-spin fa-2x"></i>
                                </span>
                                <p className="mt-3">Searching...</p>
                            </div>
                        ) : null}

                        {!loading && results.length === 0 ? (
                            <div className={`notification is-info ${isDarkMode ? 'is-dark' : 'is-light'}`} style={{ borderRadius: '8px' }}>
                                <span className="icon mr-2">
                                    <i className="fas fa-info-circle"></i>
                                </span>
                                {(searchType === 'rooms'
                                    ? "No available rooms found. Try adjusting your filters."
                                    : "No people found. Try adjusting your filters.")}
                            </div>
                        ) : searchType === 'rooms' ? (
                            <div className="table-container" style={{
                                borderRadius: '8px',
                                overflow: 'hidden',
                                opacity: loading ? 0.7 : 1,
                                transition: 'opacity 0.2s'
                            }}>
                                <table className="table is-fullwidth is-hoverable dark-mode-table">
                                    <thead>
                                        <tr>
                                            <th onClick={() => requestSort('DormName')} className={getClassNamesFor('DormName')} style={{ cursor: 'pointer', width: '30%' }}>
                                                Dorm
                                                <span className="icon is-small ml-1">
                                                    <i className={`fas ${sortConfig.key === 'DormName'
                                                        ? (sortConfig.direction === 'ascending' ? 'fa-sort-up' : 'fa-sort-down')
                                                        : 'fa-sort'}`}></i>
                                                </span>
                                            </th>
                                            <th onClick={() => requestSort('RoomID')} className={getClassNamesFor('RoomID')} style={{ cursor: 'pointer', width: '30%' }}>
                                                Room
                                                <span className="icon is-small ml-1">
                                                    <i className={`fas ${sortConfig.key === 'RoomID'
                                                        ? (sortConfig.direction === 'ascending' ? 'fa-sort-up' : 'fa-sort-down')
                                                        : 'fa-sort'}`}></i>
                                                </span>
                                            </th>
                                            <th onClick={() => requestSort('capacity')} className={getClassNamesFor('capacity')} style={{ cursor: 'pointer', width: '20%' }}>
                                                Capacity
                                                <span className="icon is-small ml-1">
                                                    <i className={`fas ${sortConfig.key === 'capacity'
                                                        ? (sortConfig.direction === 'ascending' ? 'fa-sort-up' : 'fa-sort-down')
                                                        : 'fa-sort'}`}></i>
                                                </span>
                                            </th>
                                            <th style={{ width: '20%' }}>Actions</th>
                                        </tr>
                                    </thead>
                                    <tbody>
                                        {results.map(room => (
                                            <tr key={room.RoomUUID}>
                                                <td>{room.DormName}</td>
                                                <td>{room.RoomID}</td>
                                                <td>{room.MaxOccupancy}</td>
                                                <td>
                                                    <span className="tag is-primary" style={{
                                                        borderRadius: '6px',
                                                        cursor: 'pointer',
                                                        width: '140px',
                                                        textAlign: 'center',
                                                        display: 'inline-block'
                                                    }} onClick={() => navigateToRoom(room)}>
                                                        <span className="icon is-small mr-1">
                                                            <i className="fas fa-map-marker-alt"></i>
                                                        </span>
                                                        View
                                                    </span>
                                                </td>
                                            </tr>
                                        ))}
                                    </tbody>
                                </table>
                            </div>
                        ) : (
                            <div className="table-container" style={{
                                borderRadius: '8px',
                                overflow: 'hidden',
                                opacity: loading ? 0.7 : 1,
                                transition: 'opacity 0.2s'
                            }}>
                                <table className="table is-fullwidth is-hoverable dark-mode-table">
                                    <thead>
                                        <tr>
                                            <th onClick={() => requestSort('FirstName')} className={getClassNamesFor('FirstName')} style={{ cursor: 'pointer', width: '25%' }}>
                                                Name
                                                <span className="icon is-small ml-1">
                                                    <i className={`fas ${sortConfig.key === 'FirstName'
                                                        ? (sortConfig.direction === 'ascending' ? 'fa-sort-up' : 'fa-sort-down')
                                                        : 'fa-sort'}`}></i>
                                                </span>
                                            </th>
                                            <th onClick={() => requestSort('year')} className={getClassNamesFor('year')} style={{ cursor: 'pointer', width: '20%' }}>
                                                Year
                                                <span className="icon is-small ml-1">
                                                    <i className={`fas ${sortConfig.key === 'year'
                                                        ? (sortConfig.direction === 'ascending' ? 'fa-sort-up' : 'fa-sort-down')
                                                        : 'fa-sort'}`}></i>
                                                </span>
                                            </th>
                                            <th onClick={() => requestSort('drawNumber')} className={getClassNamesFor('drawNumber')} style={{ cursor: 'pointer', width: '20%' }}>
                                                Draw Number
                                                <span className="icon is-small ml-1">
                                                    <i className={`fas ${sortConfig.key === 'drawNumber'
                                                        ? (sortConfig.direction === 'ascending' ? 'fa-sort-up' : 'fa-sort-down')
                                                        : 'fa-sort'}`}></i>
                                                </span>
                                            </th>
                                            <th style={{ width: '15%' }}>In-Dorm</th>
                                            <th style={{ width: '20%' }}>Room</th>
                                        </tr>
                                    </thead>
                                    <tbody>
                                        {results.map(user => (
                                            <tr key={user.Id}>
                                                <td>{`${user.FirstName} ${user.LastName}`}</td>
                                                <td>
                                                    <span className="tag is-info" style={{ borderRadius: '6px' }}>
                                                        {user.Year.charAt(0).toUpperCase() + user.Year.slice(1)}
                                                    </span>
                                                </td>
                                                <td>{user.DrawNumber}</td>
                                                <td>
                                                    {getDormNameFromId(user.InDorm)}
                                                </td>
                                                <td>
                                                    {user.RoomUUID &&
                                                        !user.RoomUUID.toString().match(/^0{8}-0{4}-0{4}-0{4}-0{12}$/) ? (
                                                        <span className="tag is-primary" style={{
                                                            borderRadius: '6px',
                                                            cursor: 'pointer',
                                                            width: '140px',
                                                            textAlign: 'center',
                                                            display: 'inline-block'
                                                        }} onClick={() =>
                                                        {
                                                            // Find the room details
                                                            const userRoom = rooms.find(room =>
                                                                room.RoomUUID === user.RoomUUID
                                                            );
                                                            if (userRoom)
                                                            {
                                                                navigateToRoom(userRoom);
                                                            }
                                                        }}>
                                                            <span className="icon is-small mr-1">
                                                                <i className="fas fa-map-marker-alt"></i>
                                                            </span>
                                                            View Room
                                                        </span>
                                                    ) : (
                                                        <span className="tag is-light" style={{
                                                            borderRadius: '6px',
                                                            width: '140px',
                                                            textAlign: 'center',
                                                            display: 'inline-block'
                                                        }}>
                                                            No room assigned
                                                        </span>
                                                    )}
                                                </td>
                                            </tr>
                                        ))}
                                    </tbody>
                                </table>
                            </div>
                        )}
                    </div>

                    {/* Pagination controls */}
                    {totalPages > 1 && (
                        <nav className="pagination is-centered is-small mt-4 dark-mode-pagination" role="navigation" aria-label="pagination">
                            <a
                                className="pagination-previous"
                                onClick={() =>
                                {
                                    if (page > 1 && !loading)
                                    {
                                        setLoading(true);
                                        setPage(page - 1);
                                    }
                                }}
                                disabled={page === 1 || loading}
                            >
                                Previous
                            </a>
                            <a
                                className="pagination-next"
                                onClick={() =>
                                {
                                    if (page < totalPages && !loading)
                                    {
                                        setLoading(true);
                                        setPage(page + 1);
                                    }
                                }}
                                disabled={page === totalPages || loading}
                            >
                                Next page
                            </a>
                            <ul className="pagination-list">
                                {pagination()}
                            </ul>
                        </nav>
                    )}
                </div>
            </div>
        </div>
    );
}

export default SearchPage; 