import React, { useContext, useEffect, useState } from "react";
import Select from "react-select";
import { MyContext } from "../../context/MyContext";

function SearchPage() {
    const { rooms, userMap, handleTakeMeThere, handleErrorFromTokenExpiry, isDarkMode, setCurrPage } =
        useContext(MyContext);

    const [searchType, setSearchType] = useState("rooms"); // 'rooms' or 'people'
    const [loading, setLoading] = useState(false);
    const [showLoading, setShowLoading] = useState(false); // Separate state for showing loading animation
    const [dormFilter, setDormFilter] = useState([]);
    const [capacityFilter, setCapacityFilter] = useState([]);
    const [yearFilter, setYearFilter] = useState([]);
    const [inDormFilter, setInDormFilter] = useState([]);
    const [drawNumberFilter, setDrawNumberFilter] = useState({ min: "", max: "" });
    const [results, setResults] = useState([]);
    const [showFilters, setShowFilters] = useState(true);
    const [page, setPage] = useState(1);
    const [itemsPerPage, setItemsPerPage] = useState(10);
    const [sortConfig, setSortConfig] = useState({ key: "", direction: "" });
    const [totalPages, setTotalPages] = useState(1);
    const [totalRecords, setTotalRecords] = useState(0);

    // Derived data
    const [dormOptions, setDormOptions] = useState([]);
    const [capacityOptions, setCapacityOptions] = useState([]);
    const [inDormOptions, setInDormOptions] = useState([]);

    // Add state for preplaced filter
    const [preplacedFilter, setPreplacedFilter] = useState(null);

    // Add state for hasGenderPref filter
    const [hasGenderPrefFilter, setHasGenderPrefFilter] = useState(null);

    // Add state for specific gender preference filter
    const [genderPrefFilter, setGenderPrefFilter] = useState([]);

    // Add options for preplaced filter
    const preplacedOptions = [
        { value: true, label: "Preplaced Only" },
        { value: false, label: "Non-Preplaced Only" },
    ];

    // Add options for hasGenderPref filter
    const hasGenderPrefOptions = [
        { value: true, label: "Has Gender Preferences" },
        { value: false, label: "No Gender Preferences" },
    ];

    // Add options for specific gender preference filter
    const genderPrefOptions = [
        { value: "Woman", label: "Woman" },
        { value: "Man", label: "Man" },
        { value: "Non-Binary", label: "Non-Binary" },
    ];

    // Custom styles for react-select to support dark mode
    const selectStyles = {
        control: (baseStyles) => ({
            ...baseStyles,
            backgroundColor: isDarkMode ? "var(--card-bg)" : "white",
            "&:hover": {
                borderColor: "var(--primary-color)",
            },
        }),
        option: (baseStyles, state) => ({
            ...baseStyles,
            backgroundColor: isDarkMode
                ? state.isSelected
                    ? "var(--primary-color)"
                    : state.isFocused
                      ? "var(--card-bg-hover)"
                      : "var(--card-bg)"
                : state.isSelected
                  ? "var(--primary-color)"
                  : state.isFocused
                    ? "#f5f5f5"
                    : "white",
            color: isDarkMode
                ? state.isSelected
                    ? "white"
                    : "var(--text-color)"
                : state.isSelected
                  ? "white"
                  : "inherit",
        }),
        menu: (baseStyles) => ({
            ...baseStyles,
            backgroundColor: isDarkMode ? "var(--card-bg)" : "white",
            boxShadow: "var(--box-shadow)",
        }),
        input: (baseStyles) => ({
            ...baseStyles,
            color: isDarkMode ? "var(--text-color)" : "inherit",
        }),
        singleValue: (baseStyles) => ({
            ...baseStyles,
            color: isDarkMode ? "var(--text-color)" : "inherit",
        }),
        placeholder: (baseStyles) => ({
            ...baseStyles,
            color: isDarkMode ? "var(--text-muted)" : "hsl(0, 0%, 50%)",
        }),
    };

    // Custom checkbox option component for multi-select dropdowns
    const MultiSelectOption = (props) => {
        return (
            <div {...props.innerProps} className={`multi-select-option ${props.isFocused ? "focused" : ""}`}>
                <input type="checkbox" checked={props.isSelected} onChange={() => {}} />
                <span>{props.label}</span>
            </div>
        );
    };

    // Utility function to convert dorm ID to dorm name
    const getDormNameFromId = (dormId) => {
        if (!dormId || dormId === 0) return "None";

        const dormMapping = {
            1: "East",
            2: "North",
            3: "South",
            4: "West",
            5: "Atwood",
            6: "Sontag",
            7: "Case",
            8: "Drinkward",
            9: "Linde",
        };

        return dormMapping[dormId] || `Unknown (${dormId})`;
    };

    // Utility function to simplify gender preferences display
    const simplifyGenderPreferences = (preferences) => {
        if (!preferences || !preferences.length) return [];

        const simplifiedPrefs = [...preferences];
        const hasCisWoman = simplifiedPrefs.includes("Cis Woman");
        const hasTransWoman = simplifiedPrefs.includes("Trans Woman");
        const hasCisMan = simplifiedPrefs.includes("Cis Man");
        const hasTransMan = simplifiedPrefs.includes("Trans Man");

        // Replace Cis Woman and Trans Woman with Woman if both exist
        if (hasCisWoman && hasTransWoman) {
            // Remove both individual preferences
            const indexCisWoman = simplifiedPrefs.indexOf("Cis Woman");
            simplifiedPrefs.splice(indexCisWoman, 1);

            const indexTransWoman = simplifiedPrefs.indexOf("Trans Woman");
            simplifiedPrefs.splice(indexTransWoman, 1);

            // Add combined preference
            simplifiedPrefs.push("Woman");
        }

        // Replace Cis Man and Trans Man with Man if both exist
        if (hasCisMan && hasTransMan) {
            // Remove both individual preferences
            const indexCisMan = simplifiedPrefs.indexOf("Cis Man");
            simplifiedPrefs.splice(indexCisMan, 1);

            const indexTransMan = simplifiedPrefs.indexOf("Trans Man");
            simplifiedPrefs.splice(indexTransMan, 1);

            // Add combined preference
            simplifiedPrefs.push("Man");
        }

        return simplifiedPrefs;
    };

    // Handle delayed loading indicator
    useEffect(() => {
        let timer;
        if (loading) {
            // Only show loading spinner after 1 second
            timer = setTimeout(() => {
                setShowLoading(true);
            }, 1000);
        } else {
            setShowLoading(false);
        }

        // Cleanup timer on unmount or when loading state changes
        return () => {
            if (timer) clearTimeout(timer);
        };
    }, [loading]);

    // Get unique dorms and capacities from room data
    useEffect(() => {
        if (rooms && rooms.length > 0) {
            // Extract unique dorm names
            const uniqueDorms = [...new Set(rooms.map((room) => room.DormName))];
            setDormOptions(uniqueDorms.map((dorm) => ({ value: dorm, label: dorm })));

            // Extract unique capacities
            const uniqueCapacities = [...new Set(rooms.map((room) => room.MaxOccupancy))].sort((a, b) => a - b);
            setCapacityOptions(
                uniqueCapacities.map((cap) => ({ value: cap, label: `${cap} person${cap !== 1 ? "s" : ""}` }))
            );
        }

        // Create in-dorm options from rooms
        if (rooms && rooms.length > 0) {
            const uniqueDorms = [...new Set(rooms.map((room) => room.DormName))];
            setInDormOptions(uniqueDorms.map((dorm) => ({ value: dorm, label: dorm })));
        }
    }, [rooms]);

    // Year options corrected - no "freshman" in the database
    const yearOptions = [
        { value: "sophomore", label: "Sophomore" },
        { value: "junior", label: "Junior" },
        { value: "senior", label: "Senior" },
    ];

    useEffect(() => {
        if (results.length > 0) {
            // Fetch data when page, itemsPerPage, or sorting changes
            handleSearch();
        }
    }, [page, itemsPerPage, sortConfig]);

    const handleSearch = () => {
        setLoading(true);
        // showLoading will be set to true after 1 second by the useEffect

        if (searchType === "rooms") {
            // Build query parameters for rooms search
            const params = new URLSearchParams();
            params.append("page", page);
            params.append("limit", itemsPerPage);
            params.append("empty_only", "true"); // Always filter for available rooms

            // Add dorm filter if selected
            if (dormFilter.length > 0) {
                dormFilter.forEach((dorm) => {
                    params.append("dorm", dorm.value);
                });
            }

            // Add capacity filter if selected
            if (capacityFilter.length > 0) {
                capacityFilter.forEach((capacity) => {
                    params.append("capacity", capacity.value);
                });
            }

            // Add sorting parameters
            if (sortConfig.key) {
                let serverSortKey = sortConfig.key;
                // Map frontend keys to backend keys
                const keyMapping = {
                    DormName: "dorm_name",
                    RoomID: "room_id",
                    capacity: "max_occupancy",
                    currentOccupants: "current_occupancy",
                    spacesAvailable: "max_occupancy", // Special case, we'll sort by max_occupancy
                };

                if (keyMapping[sortConfig.key]) {
                    serverSortKey = keyMapping[sortConfig.key];
                }

                params.append("sort_by", serverSortKey);
                params.append("sort_order", sortConfig.direction);
            }

            // Fetch rooms from server with pagination and sorting
            fetch(`${process.env.REACT_APP_API_URL}/search/rooms?${params.toString()}`, {
                headers: {
                    Authorization: `Bearer ${localStorage.getItem("jwt")}`,
                },
            })
                .then((response) => response.json())
                .then((data) => {
                    if (handleErrorFromTokenExpiry(data)) {
                        setLoading(false);
                        return;
                    }

                    setResults(data.rooms || []);
                    setTotalPages(data.total_pages || 1);
                    setTotalRecords(data.total || 0);
                    setLoading(false);
                })
                .catch((error) => {
                    console.error("Error fetching rooms:", error);
                    setLoading(false);
                    setResults([]);
                });
        } else {
            // Build query parameters for users search
            const params = new URLSearchParams();
            params.append("page", page);
            params.append("limit", itemsPerPage);

            // Add year filter if selected
            if (yearFilter.length > 0) {
                yearFilter.forEach((year) => {
                    params.append("year", year.value);
                });
            }

            // Add draw number range if specified
            if (drawNumberFilter.min) {
                params.append("min_draw_number", drawNumberFilter.min);
            }
            if (drawNumberFilter.max) {
                params.append("max_draw_number", drawNumberFilter.max);
            }

            // Add hasGenderPref filter if selected
            if (hasGenderPrefFilter !== null) {
                params.append("has_gender_preference", hasGenderPrefFilter.value);
            }

            // Add preplaced filter if selected
            if (preplacedFilter !== null) {
                params.append("preplaced", preplacedFilter.value);
            }

            // Add in-dorm filter if selected
            if (inDormFilter.length > 0) {
                inDormFilter.forEach((dorm) => {
                    // Get dorm ID from the dorm name
                    const dormName = dorm.value;
                    const dormRoom = rooms.find((room) => room.DormName === dormName);
                    if (dormRoom) {
                        params.append("in_dorm", dormRoom.Dorm);
                    }
                });
            }

            // Add gender preference filter if selected
            if (genderPrefFilter.length > 0) {
                genderPrefFilter.forEach((pref) => {
                    if (pref.value === "Woman") {
                        // Add both Cis Woman and Trans Woman to the query
                        params.append("gender_preference", "Cis Woman");
                        params.append("gender_preference", "Trans Woman");
                    } else if (pref.value === "Man") {
                        // Add both Cis Man and Trans Man to the query
                        params.append("gender_preference", "Cis Man");
                        params.append("gender_preference", "Trans Man");
                    } else {
                        // For other preferences like Non-Binary, add as is
                        params.append("gender_preference", pref.value);
                    }
                });
            }

            // Add sorting parameters
            if (sortConfig.key) {
                let serverSortKey = sortConfig.key;
                // Map frontend keys to backend keys
                const keyMapping = {
                    FirstName: "first_name",
                    LastName: "last_name",
                    year: "year",
                    drawNumber: "draw_number",
                };

                if (keyMapping[sortConfig.key]) {
                    serverSortKey = keyMapping[sortConfig.key];
                }

                params.append("sort_by", serverSortKey);
                params.append("sort_order", sortConfig.direction);
            }

            // Fetch users from server with pagination and sorting
            fetch(`${process.env.REACT_APP_API_URL}/search/users?${params.toString()}`, {
                headers: {
                    Authorization: `Bearer ${localStorage.getItem("jwt")}`,
                },
            })
                .then((response) => response.json())
                .then((data) => {
                    if (handleErrorFromTokenExpiry(data)) {
                        setLoading(false);
                        return;
                    }

                    setResults(data.users || []);
                    setTotalPages(data.total_pages || 1);
                    setTotalRecords(data.total || 0);
                    setLoading(false);
                })
                .catch((error) => {
                    console.error("Error fetching users:", error);
                    setLoading(false);
                    setResults([]);
                });
        }
    };

    const requestSort = (key) => {
        // Set loading state before changing sort to prevent flicker
        setLoading(true);
        // showLoading will be set to true after 1 second by the useEffect

        let direction = "asc";
        if (sortConfig.key === key && sortConfig.direction === "asc") {
            direction = "desc";
        }
        setSortConfig({ key, direction });
        setPage(1); // Reset to first page when sorting changes

        // We don't need to call handleSearch here as the useEffect will trigger it
        // This prevents double loading state
    };

    const getClassNamesFor = (name) => {
        if (!sortConfig) return;
        return sortConfig.key === name ? sortConfig.direction : "";
    };

    const resetFilters = () => {
        setDormFilter([]);
        setCapacityFilter([]);
        setYearFilter([]);
        setInDormFilter([]);
        setDrawNumberFilter({ min: "", max: "" });
        setPreplacedFilter(null);
        setHasGenderPrefFilter(null);
        setGenderPrefFilter([]);
        setResults([]);
        setSortConfig({ key: "", direction: "" });
        setPage(1);
        setTotalPages(1);
        setTotalRecords(0);
    };

    const toggleFilters = () => {
        setShowFilters(!showFilters);
    };

    const navigateToRoom = (roomInfo) => {
        const locationString = `${roomInfo.DormName} ${roomInfo.RoomID}`;
        // Switch back to Home view before navigating to room
        setCurrPage("Home");
        // Use setTimeout to ensure the home view is loaded before navigating
        setTimeout(() => {
            handleTakeMeThere(locationString, false);
        }, 50);
    };

    return (
        <div className="section">
            {/* Custom CSS for multi-select components */}
            <style>
                {`
                /* Custom Dropdown Styles */
                .react-select__control {
                    min-height: 38px !important;
                    height: auto !important;
                    display: flex !important;
                    align-items: center !important;
                    width: 100% !important;
                }
                .react-select__value-container {
                    padding: 2px 8px !important;
                    display: flex !important;
                    align-items: center !important;
                    flex-wrap: wrap !important;
                    height: auto !important;
                    overflow: visible !important;
                    width: 100% !important;
                }
                .react-select__input-container {
                    margin: 0 !important;
                    padding: 0 !important;
                }
                .react-select__placeholder {
                    margin: 0 2px !important;
                }
                .react-select__multi-value {
                    margin: 2px 4px 2px 0 !important;
                    display: flex !important;
                    align-items: center !important;
                }
                .react-select__indicators {
                    align-self: center !important;
                    height: 100% !important;
                    display: flex !important;
                    align-items: center !important;
                }
                .react-select__menu {
                    width: 100% !important;
                    z-index: 20 !important;
                }
                .react-select__option {
                    padding: 8px 12px !important;
                    display: flex !important;
                    align-items: center !important;
                }
                .react-select__single-value {
                    display: flex !important;
                    align-items: center !important;
                    margin: 0 2px !important;
                }
                
                /* Container for react-select */
                .react-select {
                    width: 100% !important;
                }
                
                .react-select > div {
                    width: 100% !important;
                }
                
                /* Custom MultiSelectOption styling */
                .multi-select-option {
                    padding: 8px 12px;
                    cursor: pointer;
                    display: flex;
                    align-items: center;
                    background-color: white;
                }
                
                .multi-select-option.focused {
                    background-color: #f5f5f5;
                }
                
                .multi-select-option input[type="checkbox"] {
                    margin-right: 8px;
                }
                
                body.dark-mode .multi-select-option {
                    background-color: var(--card-bg);
                    color: var(--text-color);
                }
                
                body.dark-mode .multi-select-option.focused {
                    background-color: var(--card-bg-hover);
                }
                `}
            </style>
            <div className="container">
                <h1 className="title has-text-centered mb-5">Search</h1>

                <div className="box mb-4 search-box">
                    <div className="buttons has-addons is-centered mb-3">
                        <button
                            className={`button ${searchType === "rooms" ? "is-primary" : isDarkMode ? "is-dark" : "is-light"}`}
                            onClick={() => {
                                setSearchType("rooms");
                                resetFilters();
                            }}
                            style={{ borderRadius: "8px 0 0 8px", transition: "all 0.3s ease" }}
                        >
                            <span className="icon">
                                <i className="fas fa-door-open"></i>
                            </span>
                            <span>Find Empty Rooms</span>
                        </button>
                        <button
                            className={`button ${searchType === "people" ? "is-primary" : isDarkMode ? "is-dark" : "is-light"}`}
                            onClick={() => {
                                setSearchType("people");
                                resetFilters();
                            }}
                            style={{ borderRadius: "0 8px 8px 0", transition: "all 0.3s ease" }}
                        >
                            <span className="icon">
                                <i className="fas fa-user"></i>
                            </span>
                            <span>Find People</span>
                        </button>
                    </div>

                    <div className="is-flex is-justify-content-flex-end mb-3">
                        <button
                            className={`button is-small ${isDarkMode ? "is-dark" : "is-light"}`}
                            onClick={toggleFilters}
                            title={showFilters ? "Hide filters" : "Show filters"}
                            style={{ borderRadius: "6px", transition: "all 0.3s ease" }}
                        >
                            <span className="icon">
                                <i className={`fas ${showFilters ? "fa-chevron-up" : "fa-chevron-down"}`}></i>
                            </span>
                            <span>{showFilters ? "Hide Filters" : "Show Filters"}</span>
                        </button>
                    </div>

                    {showFilters && (
                        <div className="px-4 py-3 search-filter-panel">
                            {searchType === "rooms" ? (
                                <div className="columns is-centered">
                                    <div
                                        className="column is-5"
                                        style={{ display: "flex", flexDirection: "column", height: "100px" }}
                                    >
                                        <label className="label dark-mode-label">Dorm</label>
                                        <div
                                            style={{ flex: 1, display: "flex", flexDirection: "column", width: "100%" }}
                                        >
                                            <Select
                                                className="react-select"
                                                classNamePrefix="react-select"
                                                placeholder="All Dorms"
                                                options={dormOptions}
                                                value={dormFilter}
                                                onChange={(options) => setDormFilter(options || [])}
                                                isClearable
                                                isMulti
                                                closeMenuOnSelect={false}
                                                hideSelectedOptions={false}
                                                components={{ Option: MultiSelectOption }}
                                                styles={{
                                                    ...selectStyles,
                                                }}
                                            />
                                        </div>
                                    </div>
                                    <div
                                        className="column is-5"
                                        style={{ display: "flex", flexDirection: "column", height: "100px" }}
                                    >
                                        <label className="label dark-mode-label">Room Capacity</label>
                                        <div
                                            style={{ flex: 1, display: "flex", flexDirection: "column", width: "100%" }}
                                        >
                                            <Select
                                                className="react-select"
                                                classNamePrefix="react-select"
                                                placeholder="Any Size"
                                                options={capacityOptions}
                                                value={capacityFilter}
                                                onChange={(options) => setCapacityFilter(options || [])}
                                                isClearable
                                                isMulti
                                                closeMenuOnSelect={false}
                                                hideSelectedOptions={false}
                                                components={{ Option: MultiSelectOption }}
                                                styles={{
                                                    ...selectStyles,
                                                }}
                                            />
                                        </div>
                                    </div>
                                </div>
                            ) : (
                                <>
                                    <div className="columns mb-3" style={{ alignItems: "flex-start", display: "flex" }}>
                                        <div
                                            className="column is-3"
                                            style={{ display: "flex", flexDirection: "column", height: "100px" }}
                                        >
                                            <label className="label dark-mode-label">Year</label>
                                            <div
                                                style={{
                                                    flex: 1,
                                                    display: "flex",
                                                    flexDirection: "column",
                                                    width: "100%",
                                                }}
                                            >
                                                <Select
                                                    className="react-select"
                                                    classNamePrefix="react-select"
                                                    placeholder="All Years"
                                                    options={yearOptions}
                                                    value={yearFilter}
                                                    onChange={(options) => setYearFilter(options || [])}
                                                    isClearable
                                                    isMulti
                                                    closeMenuOnSelect={false}
                                                    hideSelectedOptions={false}
                                                    components={{ Option: MultiSelectOption }}
                                                    styles={{
                                                        ...selectStyles,
                                                    }}
                                                />
                                            </div>
                                        </div>
                                        <div
                                            className="column is-3"
                                            style={{ display: "flex", flexDirection: "column", height: "100px" }}
                                        >
                                            <label className="label dark-mode-label">Has Gender Preference</label>
                                            <div
                                                style={{
                                                    flex: 1,
                                                    display: "flex",
                                                    flexDirection: "column",
                                                    width: "100%",
                                                }}
                                            >
                                                <Select
                                                    className="react-select"
                                                    classNamePrefix="react-select"
                                                    placeholder="Any"
                                                    options={hasGenderPrefOptions}
                                                    value={hasGenderPrefFilter}
                                                    onChange={(option) => setHasGenderPrefFilter(option)}
                                                    isClearable
                                                    styles={{
                                                        ...selectStyles,
                                                    }}
                                                />
                                            </div>
                                        </div>
                                        <div
                                            className="column is-3"
                                            style={{ display: "flex", flexDirection: "column", height: "100px" }}
                                        >
                                            <label className="label dark-mode-label">Gender Preference</label>
                                            <div
                                                style={{
                                                    flex: 1,
                                                    display: "flex",
                                                    flexDirection: "column",
                                                    width: "100%",
                                                }}
                                            >
                                                <Select
                                                    className="react-select"
                                                    classNamePrefix="react-select"
                                                    placeholder="Any Gender"
                                                    options={genderPrefOptions}
                                                    value={genderPrefFilter}
                                                    onChange={(options) => setGenderPrefFilter(options || [])}
                                                    isClearable
                                                    isMulti
                                                    closeMenuOnSelect={false}
                                                    hideSelectedOptions={false}
                                                    components={{ Option: MultiSelectOption }}
                                                    styles={{
                                                        ...selectStyles,
                                                    }}
                                                />
                                            </div>
                                        </div>
                                        <div
                                            className="column is-3"
                                            style={{ display: "flex", flexDirection: "column", height: "100px" }}
                                        >
                                            <label className="label dark-mode-label">Preplaced Status</label>
                                            <div
                                                style={{
                                                    flex: 1,
                                                    display: "flex",
                                                    flexDirection: "column",
                                                    width: "100%",
                                                }}
                                            >
                                                <Select
                                                    className="react-select"
                                                    classNamePrefix="react-select"
                                                    placeholder="Any Status"
                                                    options={preplacedOptions}
                                                    value={preplacedFilter}
                                                    onChange={(option) => setPreplacedFilter(option)}
                                                    isClearable
                                                    styles={{
                                                        ...selectStyles,
                                                    }}
                                                />
                                            </div>
                                        </div>
                                    </div>
                                    <div className="columns">
                                        <div className="column is-6 is-offset-3">
                                            <label className="label dark-mode-label">Draw Number Range</label>
                                            <div className="field has-addons">
                                                <div className="control is-expanded">
                                                    <input
                                                        className="input"
                                                        type="number"
                                                        placeholder="Min"
                                                        min="1"
                                                        value={drawNumberFilter.min}
                                                        onChange={(e) =>
                                                            setDrawNumberFilter({
                                                                ...drawNumberFilter,
                                                                min: e.target.value,
                                                            })
                                                        }
                                                    />
                                                </div>
                                                <div className="control">
                                                    <span className="button is-static">to</span>
                                                </div>
                                                <div className="control is-expanded">
                                                    <input
                                                        className="input"
                                                        type="number"
                                                        placeholder="Max"
                                                        min="1"
                                                        value={drawNumberFilter.max}
                                                        onChange={(e) =>
                                                            setDrawNumberFilter({
                                                                ...drawNumberFilter,
                                                                max: e.target.value,
                                                            })
                                                        }
                                                    />
                                                </div>
                                            </div>
                                        </div>
                                    </div>
                                </>
                            )}

                            <div className="field is-grouped is-flex is-justify-content-center mt-4">
                                <div className="control">
                                    <button
                                        className={`button is-primary ${loading ? "is-loading" : ""}`}
                                        onClick={() => {
                                            setPage(1); // Reset to first page when searching
                                            handleSearch();
                                        }}
                                        style={{ borderRadius: "8px", transition: "all 0.3s ease", width: "120px" }}
                                    >
                                        <span className="icon">
                                            <i className="fas fa-search"></i>
                                        </span>
                                        <span>Search</span>
                                    </button>
                                </div>
                                <div className="control">
                                    <button
                                        className={`button ${isDarkMode ? "is-dark" : "is-light"}`}
                                        onClick={resetFilters}
                                        style={{ borderRadius: "8px", transition: "all 0.3s ease" }}
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
                                        <i className={`fas ${searchType === "rooms" ? "fa-door-open" : "fa-user"}`}></i>
                                    </span>
                                    {searchType === "rooms" ? "Available Rooms" : "People"}
                                </h3>
                            </div>
                        </div>
                        <div className="level-right">
                            <div className="level-item">
                                <div className="is-flex is-align-items-center">
                                    <div className="select is-small mr-2">
                                        <select
                                            value={itemsPerPage}
                                            onChange={(e) => {
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

                    <div className="results-content" style={{ minHeight: "300px", position: "relative" }}>
                        {loading && showLoading ? (
                            <div className="has-text-centered search-loading-overlay">
                                <span className="icon is-large">
                                    <i className="fas fa-circle-notch fa-spin fa-2x"></i>
                                </span>
                                <p className="mt-3">Searching...</p>
                            </div>
                        ) : null}

                        {!loading && results.length === 0 ? (
                            <div
                                className={`notification is-info ${isDarkMode ? "is-dark" : "is-light"}`}
                                style={{ borderRadius: "8px" }}
                            >
                                <span className="icon mr-2">
                                    <i className="fas fa-info-circle"></i>
                                </span>
                                {searchType === "rooms"
                                    ? "No available rooms found. Try adjusting your filters."
                                    : "No people found. Try adjusting your filters."}
                            </div>
                        ) : searchType === "rooms" ? (
                            <div
                                className="table-container"
                                style={{
                                    borderRadius: "8px",
                                    overflow: "hidden",
                                    opacity: loading ? 0.7 : 1,
                                    transition: "opacity 0.2s",
                                }}
                            >
                                <table className="table is-fullwidth is-hoverable dark-mode-table">
                                    <thead>
                                        <tr>
                                            <th
                                                onClick={() => requestSort("DormName")}
                                                className={getClassNamesFor("DormName")}
                                                style={{ cursor: "pointer", width: "30%" }}
                                            >
                                                Dorm
                                                <span className="icon is-small ml-1">
                                                    <i
                                                        className={`fas ${
                                                            sortConfig.key === "DormName"
                                                                ? sortConfig.direction === "asc"
                                                                    ? "fa-sort-up"
                                                                    : "fa-sort-down"
                                                                : "fa-sort"
                                                        }`}
                                                    ></i>
                                                </span>
                                            </th>
                                            <th
                                                onClick={() => requestSort("RoomID")}
                                                className={getClassNamesFor("RoomID")}
                                                style={{ cursor: "pointer", width: "30%" }}
                                            >
                                                Room
                                                <span className="icon is-small ml-1">
                                                    <i
                                                        className={`fas ${
                                                            sortConfig.key === "RoomID"
                                                                ? sortConfig.direction === "asc"
                                                                    ? "fa-sort-up"
                                                                    : "fa-sort-down"
                                                                : "fa-sort"
                                                        }`}
                                                    ></i>
                                                </span>
                                            </th>
                                            <th
                                                onClick={() => requestSort("capacity")}
                                                className={getClassNamesFor("capacity")}
                                                style={{ cursor: "pointer", width: "20%" }}
                                            >
                                                Capacity
                                                <span className="icon is-small ml-1">
                                                    <i
                                                        className={`fas ${
                                                            sortConfig.key === "capacity"
                                                                ? sortConfig.direction === "asc"
                                                                    ? "fa-sort-up"
                                                                    : "fa-sort-down"
                                                                : "fa-sort"
                                                        }`}
                                                    ></i>
                                                </span>
                                            </th>
                                            <th style={{ width: "20%" }}>Actions</th>
                                        </tr>
                                    </thead>
                                    <tbody>
                                        {results.map((room) => (
                                            <tr key={room.RoomUUID}>
                                                <td>{room.DormName}</td>
                                                <td>{room.RoomID}</td>
                                                <td>{room.MaxOccupancy}</td>
                                                <td>
                                                    <span
                                                        className="tag is-primary"
                                                        style={{
                                                            borderRadius: "6px",
                                                            cursor: "pointer",
                                                            width: "140px",
                                                            textAlign: "center",
                                                            display: "inline-block",
                                                        }}
                                                        onClick={() => navigateToRoom(room)}
                                                    >
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
                            <div
                                className="table-container"
                                style={{
                                    borderRadius: "8px",
                                    overflow: "hidden",
                                    opacity: loading ? 0.7 : 1,
                                    transition: "opacity 0.2s",
                                }}
                            >
                                <table className="table is-fullwidth is-hoverable dark-mode-table">
                                    <thead>
                                        <tr>
                                            <th
                                                onClick={() => requestSort("FirstName")}
                                                className={getClassNamesFor("FirstName")}
                                                style={{ cursor: "pointer", width: "20%" }}
                                            >
                                                Name
                                                <span className="icon is-small ml-1">
                                                    <i
                                                        className={`fas ${
                                                            sortConfig.key === "FirstName"
                                                                ? sortConfig.direction === "asc"
                                                                    ? "fa-sort-up"
                                                                    : "fa-sort-down"
                                                                : "fa-sort"
                                                        }`}
                                                    ></i>
                                                </span>
                                            </th>
                                            <th
                                                onClick={() => requestSort("year")}
                                                className={getClassNamesFor("year")}
                                                style={{ cursor: "pointer", width: "15%" }}
                                            >
                                                Year
                                                <span className="icon is-small ml-1">
                                                    <i
                                                        className={`fas ${
                                                            sortConfig.key === "year"
                                                                ? sortConfig.direction === "asc"
                                                                    ? "fa-sort-up"
                                                                    : "fa-sort-down"
                                                                : "fa-sort"
                                                        }`}
                                                    ></i>
                                                </span>
                                            </th>
                                            <th
                                                onClick={() => requestSort("drawNumber")}
                                                className={getClassNamesFor("drawNumber")}
                                                style={{ cursor: "pointer", width: "15%" }}
                                            >
                                                Draw Number
                                                <span className="icon is-small ml-1">
                                                    <i
                                                        className={`fas ${
                                                            sortConfig.key === "drawNumber"
                                                                ? sortConfig.direction === "asc"
                                                                    ? "fa-sort-up"
                                                                    : "fa-sort-down"
                                                                : "fa-sort"
                                                        }`}
                                                    ></i>
                                                </span>
                                            </th>
                                            <th style={{ width: "15%" }}>Gender Preferences</th>
                                            <th style={{ width: "10%" }}>In-Dorm</th>
                                            <th style={{ width: "15%" }}>Room</th>
                                        </tr>
                                    </thead>
                                    <tbody>
                                        {results.map((user) => (
                                            <tr key={user.Id}>
                                                <td>{`${user.FirstName} ${user.LastName}`}</td>
                                                <td>
                                                    <span className="tag is-info" style={{ borderRadius: "6px" }}>
                                                        {user.Year.charAt(0).toUpperCase() + user.Year.slice(1)}
                                                    </span>
                                                </td>
                                                <td>
                                                    {user.Preplaced ? (
                                                        <span
                                                            className="tag is-success"
                                                            style={{ borderRadius: "6px" }}
                                                        >
                                                            Preplaced
                                                        </span>
                                                    ) : (
                                                        user.DrawNumber
                                                    )}
                                                </td>
                                                <td>
                                                    {user.GenderPreferences && user.GenderPreferences.length > 0 ? (
                                                        <div className="tags">
                                                            {simplifyGenderPreferences(user.GenderPreferences).map(
                                                                (pref, index) => (
                                                                    <span
                                                                        key={index}
                                                                        className="tag is-primary"
                                                                        style={{ borderRadius: "6px" }}
                                                                    >
                                                                        {pref}
                                                                    </span>
                                                                )
                                                            )}
                                                        </div>
                                                    ) : (
                                                        <span className="tag is-light" style={{ borderRadius: "6px" }}>
                                                            None
                                                        </span>
                                                    )}
                                                </td>
                                                <td>{getDormNameFromId(user.InDorm)}</td>
                                                <td>
                                                    {user.RoomUUID &&
                                                    !user.RoomUUID.toString().match(/^0{8}-0{4}-0{4}-0{4}-0{12}$/) ? (
                                                        <span
                                                            className="tag is-primary"
                                                            style={{
                                                                borderRadius: "6px",
                                                                cursor: "pointer",
                                                                width: "140px",
                                                                textAlign: "center",
                                                                display: "inline-block",
                                                            }}
                                                            onClick={() => {
                                                                // Find the room details
                                                                const userRoom = rooms.find(
                                                                    (room) => room.RoomUUID === user.RoomUUID
                                                                );
                                                                if (userRoom) {
                                                                    navigateToRoom(userRoom);
                                                                }
                                                            }}
                                                        >
                                                            <span className="icon is-small mr-1">
                                                                <i className="fas fa-map-marker-alt"></i>
                                                            </span>
                                                            View Room
                                                        </span>
                                                    ) : (
                                                        <span
                                                            className="tag is-light"
                                                            style={{
                                                                borderRadius: "6px",
                                                                width: "140px",
                                                                textAlign: "center",
                                                                display: "inline-block",
                                                            }}
                                                        >
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
                        <nav className="pagination is-centered" role="navigation" aria-label="pagination">
                            <button
                                className="pagination-previous"
                                onClick={() => {
                                    if (page > 1 && !loading) {
                                        setLoading(true);
                                        setPage(page - 1);
                                    }
                                }}
                                disabled={page === 1 || loading}
                            >
                                <span className="icon">
                                    <i className="fas fa-chevron-left"></i>
                                </span>
                                <span>Previous</span>
                            </button>
                            <button
                                className="pagination-next"
                                onClick={() => {
                                    if (page < totalPages && !loading) {
                                        setLoading(true);
                                        setPage(page + 1);
                                    }
                                }}
                                disabled={page === totalPages || loading}
                            >
                                <span>Next</span>
                                <span className="icon">
                                    <i className="fas fa-chevron-right"></i>
                                </span>
                            </button>
                            <ul className="pagination-list">
                                {/* First page */}
                                {page > 2 && (
                                    <li>
                                        <button className="pagination-link" onClick={() => setPage(1)}>
                                            1
                                        </button>
                                    </li>
                                )}

                                {/* Ellipsis if needed */}
                                {page > 3 && (
                                    <li>
                                        <span className="pagination-ellipsis">&hellip;</span>
                                    </li>
                                )}

                                {/* Previous page if not first */}
                                {page > 1 && (
                                    <li>
                                        <button className="pagination-link" onClick={() => setPage(page - 1)}>
                                            {page - 1}
                                        </button>
                                    </li>
                                )}

                                {/* Current page */}
                                <li>
                                    <button className="pagination-link is-current" aria-current="page">
                                        {page}
                                    </button>
                                </li>

                                {/* Next page if not last */}
                                {page < totalPages && (
                                    <li>
                                        <button className="pagination-link" onClick={() => setPage(page + 1)}>
                                            {page + 1}
                                        </button>
                                    </li>
                                )}

                                {/* Ellipsis if needed */}
                                {page < totalPages - 2 && (
                                    <li>
                                        <span className="pagination-ellipsis">&hellip;</span>
                                    </li>
                                )}

                                {/* Last page */}
                                {page < totalPages - 1 && (
                                    <li>
                                        <button className="pagination-link" onClick={() => setPage(totalPages)}>
                                            {totalPages}
                                        </button>
                                    </li>
                                )}
                            </ul>
                        </nav>
                    )}
                </div>
            </div>
        </div>
    );
}

export default SearchPage;
