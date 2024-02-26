import React, { useEffect, useState, useContext } from 'react';
import 'bulma/css/bulma.min.css';
import BumpModal from './BumpModal';
import FloorGrid from './FloorGrid';
import Recommendations from './Recommendations';
import { MyContext } from './MyContext';
import { GoogleLogin } from '@react-oauth/google';
import { jwtDecode } from "jwt-decode";
import  SuiteNoteModal  from './SuiteNoteModal';
import { googleLogout } from '@react-oauth/google';

function App() {
  const {
    currPage,
    setCurrPage,
    gridData,
    userMap,
    isModalOpen,
    dormMapping,
    onlyShowBumpableRooms,
    setOnlyShowBumpableRooms,
    getNameById,
    selectedID,
    setSelectedID,
    rooms,
    isSuiteNoteModalOpen,
    credentials,
    setCredentials,
    lastRefreshedTime,
    activeTab,
    setActiveTab

  } = useContext(MyContext);

  // const [showNotification, setShowNotification] = useState(false);
  const [myRoom, setMyRoom] = useState("You are not in a room yet."); // to show what room current user is in

  useEffect(() => {
    const storedCredentials = localStorage.getItem('jwt');
    if (storedCredentials) {
      console.log("use effect");
      console.log(storedCredentials);
      console.log("end use efect");
      setCredentials(storedCredentials);
    }
  }, []);

  // useEffect(() => {
  //   // Check for stored credentials on component mount
  //   const storedCredentials = localStorage.getItem('jwt');
  //   if (storedCredentials) {
  //     // If credentials exist, decode and set them
  //     setCredentials(storedCredentials);
  //   }
  // }, []);

  const handleSuccess = (credentialResponse) => {
    // decode the credential

    // const decoded = jwtDecode(credentialResponse.credential);
    // console.log(decoded.given_name);

    setCredentials(credentialResponse.credential);
    localStorage.setItem('jwt', credentialResponse.credential); // originally stored credentialREsponse
    // localStorage.setItem('jwt', credentials); // originally stored credentialREsponse
    console.log(typeof credentialResponse.credential);
  };

  const handleError = () => {
    console.log('Login Failed');
    // Optionally, handle login failure (e.g., by clearing stored credentials)
  };

  const handleLogout = () => {
    setCredentials(null);
    localStorage.removeItem('jwt');
    googleLogout();
  };

  useEffect(() => {
    // updates room that thei current user is in every time the selected user or the room data changes
    if (!rooms || !Array.isArray(rooms)) {
      return;
    }
    if (rooms) {
      for (let room of rooms) {

        if (room.Occupants && room.Occupants.includes(Number(selectedID))) {


          setMyRoom(`You are in ${room.DormName} ${room.RoomID}.`);
          return;
        }
      }
      setMyRoom("You are not in a room yet.");


    }
  }, [selectedID, rooms]);



  // Save state to localStorage whenever it changes
  useEffect(() => {
    localStorage.setItem('activeTab', activeTab);
  }, [activeTab]);


  const handleNameChange = (event) => {
    setSelectedID(event.target.value);
  };

  const handleTabClick = (tab) => {
    setActiveTab(tab);
  };

  function getDrawNumberAndYear(id) {
    // Find the drawNumber in laymans terms with the given id, including in-dorm status
    // ex: given 2, returns Sophomore 46
    if (!userMap) {
      return "Loading...";
    } else if (userMap[id].InDorm && userMap[id].InDorm !== 0) {
      // has in dorm
      return `${userMap[id].Year.charAt(0).toUpperCase() + userMap[id].Year.slice(1)} ${userMap[id].DrawNumber} ${dormMapping[userMap[id].InDorm]}`;
    }
    return `${userMap[id].Year.charAt(0).toUpperCase() + userMap[id].Year.slice(1)} ${userMap[id].DrawNumber}`
  }


  // Component for each floor, to show even and odd floors separately
  const FloorColumn = ({ gridData, filterCondition }) => (
    <div class="column">
      {gridData.map((dorm) => (
        <div key={dorm.dormName} className={activeTab === dorm.dormName ? '' : 'is-hidden'}>
          {dorm.floors
            .filter((floor) => filterCondition(floor.floorNumber))
            .sort((a, b) => Number(a.floorNumber) - Number(b.floorNumber))  // Convert to numbers before comparing
            .map((floor, floorIndex) => (
              <div style={{ paddingBottom: 20 }} className="container" key={floorIndex}>
                <h2 class="subtitle has-text-centered">Floor {floor.floorNumber + 1}</h2>
                <FloorGrid gridData={floor} />
              </div>
            ))}
        </div>
      ))}
    </div>
  );
  return (
    <div>

      <nav class="navbar" role="navigation" aria-label="main navigation">
        <div class="navbar-brand">
          <a class="navbar-item" href="https://ibb.co/c3D21bJ"><img src="https://i.ibb.co/SyRVPQN/Screenshot-2023-12-26-at-10-14-31-PM.png" alt="Screenshot-2023-12-26-at-10-14-31-PM" border="0" /></a>

          <a role="button" class="navbar-burger" aria-label="menu" aria-expanded="false" data-target="navbarBasicExample">
            <span aria-hidden="true"></span>
            <span aria-hidden="true"></span>
            <span aria-hidden="true"></span>
          </a>
        </div>

        <div id="navbarBasicExample" class="navbar-menu">
          <div class="navbar-start">
            <a class="navbar-item" onClick={() => setCurrPage("Home")}>
              Home
            </a>

            <a class="navbar-item" onClick={() => setCurrPage("Recommendations")}>
              Recommendations
            </a>

            <div class="navbar-item has-dropdown is-hoverable">
              <a class="navbar-link">
                More
              </a>

              <div class="navbar-dropdown">
                <a class="navbar-item">
                  About
                </a>
                <a class="navbar-item">
                  Jobs
                </a>
                <a class="navbar-item">
                  Contact
                </a>

                <a class="navbar-item">
                  Report an issue
                </a>
              </div>
            </div>
          </div>

          <div class="navbar-end">
            <div class="navbar-item">
              <div class="buttons">
                {!credentials && 
                  <GoogleLogin auto_select={true}
                  onSuccess={handleSuccess}
                  onError={handleError}
                />}
                {credentials && <a class="button is-secondary">
                  <strong>Welcome, {jwtDecode(credentials).given_name} </strong> 
                </a>}
                {credentials && <a class="button is-danger" onClick={handleLogout}>
                  <strong>Log Out</strong> 
                </a>}
              </div>
            </div>
          </div>
        </div>
      </nav>
      {isModalOpen && <BumpModal />}
      {isSuiteNoteModalOpen && <SuiteNoteModal />}

    {!credentials && <section class="section">
      <div style={{ textAlign: 'center' }}>
        <h1 className="title">Welcome to Digital Draw!</h1>
        <h2 className="subtitle">Please log in with your HMC email to continue.</h2>
      </div>
    </section>}
    {credentials && <section class="section">
      <div style={{ textAlign: 'center' }}>

        <h1 className="title">You're viewing DigiDraw as {getNameById(selectedID)}. <br /> </h1>
        <h2 className="subtitle">You are <strong>{getDrawNumberAndYear(selectedID)}</strong>. {myRoom} <br />Click on any room you'd like to change! <br/>Last refreshed at {lastRefreshedTime.toLocaleTimeString()}.</h2>

        <div style={{ display: 'flex', justifyContent: 'center' }}>
          <select className="select" onChange={handleNameChange}>
            <option value="">This isn't me</option>
            {userMap && Object.keys(userMap)
              .sort((a, b) => {
                const nameA = `${userMap[a].FirstName} ${userMap[a].LastName}`.toUpperCase();
                const nameB = `${userMap[b].FirstName} ${userMap[b].LastName}`.toUpperCase();
                if (nameA < nameB) {
                  return -1;
                }
                if (nameA > nameB) {
                  return 1;
                }
                return 0;
              })
              .map((key, index) => (
                <option key={index} value={key}>
                  {userMap[key].FirstName} {userMap[key].LastName}
                </option>
              ))}
          </select>
        </div>

      </div>

    </section>}

    {(credentials && currPage === "Home") && <section class="section">
      <label className="checkbox" style={{ display: 'flex', alignItems: 'center' }}>
        <input
          type="checkbox"
          checked={onlyShowBumpableRooms}
          onChange={() => setOnlyShowBumpableRooms(!onlyShowBumpableRooms)}
        />
        <span style={{ marginLeft: '0.5rem' }}>Darken rooms I can't pull</span>
      </label>

      <div className="tabs is-centered">
        <ul>

          {gridData.map((dorm) => (
            <li
              key={dorm.dormName}
              className={activeTab === dorm.dormName ? 'is-active' : ''}
              onClick={() => handleTabClick(dorm.dormName)}
            >

              <a>{dorm.dormName}</a>
            </li>
          ))}
        </ul>
      </div>

      {/* Left column is room draw, right side is tips */}
      <div class="columns">

        <FloorColumn gridData={gridData} filterCondition={(floorNumber) => floorNumber === 0} />
        <FloorColumn gridData={gridData} filterCondition={(floorNumber) => floorNumber === 1} />
        <FloorColumn gridData={gridData} filterCondition={(floorNumber) => floorNumber === 2} />


      </div>


    </section>}
      {currPage === "Recommendations" && <section class="section">
        {/* <Recommendations gridData={gridData} setCurrPage={setCurrPage} /> */}
      </section>}


      <footer class="footer">
        <div class="content has-text-centered">
          <p>
            <strong>Digital Draw</strong> by Serena Mao & Tom Lam. Email smao@g.hmc.edu or tlam@g.hmc.edu with questions.
          </p>
        </div>
      </footer>


    </div>

  );
}


export default App;