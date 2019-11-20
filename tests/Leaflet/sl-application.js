
/**
 * Namespace maker
 *
 * by Ringo
 *
 * A very simple way to create a jquery namespace to help organize related
 * methods, constants, etc.
 */


if (typeof($) === 'undefined') {
    $ = {};
}

/**
 * Define and construct the jquery namespace
 *
 * @return {null}
 */
$.namespace = function() {
    var o = null;
    var i, j, d;
    for (i = 0; i < arguments.length; i++) {
        d = arguments[i].split(".");
        o = window;
        for (j = 0; j < d.length; j++) {
            o[ d[ j ] ] = o[ d[ j ] ] || {};
            o = o[ d[ j ] ];
        }
    }
    return o;
};

// Namespace declarations
// Add new declarations here!
$.namespace('$.sl.maps.config');
$.namespace('$.sl.maps');
$.sl.maps.config = {
    sl_base_url: "https://secondlife.com/",
    // tile_url: "https://secondlife-maps-cdn.akamaized.net",
    tile_url: "http://opensim.betatechnologies.info:8002",

    //Default Destination Information
    default_title: "Welcome to Second Life",
    default_img: "https://secondlife-maps-lecs.akamaized.net/agni/default-new.jpg",
    default_msg: "Second Life is a popular virtual space for meeting friends, doing business, and sharing knowledge. If you have Second Life installed on your computer, teleport in and start exploring!",

    // Turn on the map debugger
    map_debug: false,

    // The maximum width/height of the SL grid in regions:
    // 2^20 regions on a side = 1,048,786    ("This should be enough for anyone")
    // *NOTE: This must be a power of 2 and divisible by 2^(max zoom) = 256
    map_grid_edge_size: 1048576,

    // NOTE: Mustachejs Templates
    // These templates were moved in to here so they are available as soon as the javascript is loaded
    // that way we don't have to wait for the page to full load before processing all the data we have on hand.

    // Ballooon Popup Template
    balloon_tmpl: ' \
        <div class="balloon-content"> \
            <h3> \
                {{#slurl}} \
                <a href="{{slurl}}" onclick="trakkit(\'maps\', \'teleport\', \'{{slurl}}\');"> \
                {{/slurl}} \
                    {{title}} \
                {{#slurl}} \
                </a> \
                {{/slurl}} \
            </h3> \
            {{#img}} \
            <a href="{{slurl}}" onclick="trakkit(\'maps\', \'teleport\', \'{{slurl}}\');"> \
                <img src="{{img}}" onError="this.onerror=null;this.src=assetsURL + \'default-new.jpg\';" /> \
            </a> \
            {{/img}} \
            <p>{{msg}}</p> \
            <div class="buttons"> \
                {{#slurl}} \
                    <a class="HIGHLANDER_button_hot btn_large primary" title="visit this location" href="{{slurl}}" onclick="trakkit(\'maps\', \'teleport\', \'{{slurl}}\');">Visit this location</a> \
                {{/slurl}} \
                <a href="https://join.secondlife.com/" target="_top" class="HIGHLANDER_button_hot btn_large secondary join_button">Join Now, it&rsquo;s free!</a> \
            </div> \
        </div>',

    // Slurl Doesn't Exist Templte
    noexists_tmpl: ' \
        <div id="map-error"> \
            <div id="error-content"> \
                <span class="error-close">Hide message</span> \
                <span class="location-title">We are unable to locate the region "{{region_name}}"</span> \
                <p>This region may no longer exist, but please double check your spelling and coordinates to make sure there aren&rsquo;t any errors and try again.</p> \
                <p>If your problem persists, contact <a href="http://secondlife.com/support/">Second Life support</a></p> \
            </div> \
        </div>'
}
;
// License and Terms of Use
//
// Copyright 2016 Linden Research, Inc.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.
//
// This javascript makes use of the Second Life Map API, which is documented
// at http://wiki.secondlife.com/wiki/Map_API
//
// Use of the Second Life Map API is subject to the Second Life API Terms of Use:
//   https://wiki.secondlife.com/wiki/Linden_Lab_Official:API_Terms_of_Use
//
// Questions regarding this javascript, and any suggested improvements to it,
// should be sent to the mailing list opensource-dev@list.secondlife.com
// ==============================================================================

/*

# How A Leaflet Map #

Unlike Google Maps, Leaflet provides a "Simple" coordinate space for 2D maps
like Second Life's out of the box. The Simple CRS is not even limited by a
geographic conception of coordinate limits: we can really use the real
regionspace coordinates & the Simple CRS operates correctly. This means we have
nonsensical (Earth-wise) "latitudes" & "longitudes", because they're really
just coordinates.

<http://leafletjs.com/examples/crs-simple/crs-simple.html>

For further simplicity, the possible Second Life region space (labeled here
"grid_x, grid_y") is mapped to the upper right quadrant of such a map. This
puts the (0, 0) origin of region space at LatLng (0, 0).

    (0, 2^20) long, lat
    (0, 2^20) grid_x, grid_y
     |
     V
     X------------------------------------+
     |                                    |
     |                                    |
     |                                    |
     |                                    |
     |                                    |
     |                                    |
     |                                    |
     |                                    |
     | xxx                                |
     | xxx                                |   (2^20, 0) long, lat
     X------------------------------------X<- (2^20, 0) grid_x, grid_y
     ^
     |
    (0, 0) long, lat
    (0, 0) grid_x, grid_y

A large scaling value called `map_grid_edge_size` defines the largest region
coordinate at the top and far right edge of the map. At the current value of
2^20 = 1M, this creates a map area with room for 1 trillion regions. The xxx'd
area of the map represents today's populated regions.

Previously, Google Maps API v3 mapped maps into a 0-256 "world coordinate" box,
so we just used that, putting the 0 edge of regionspace at -180 longitude.
Google Maps API v2 before it still used lat/lng as Leaflet does, so we are
returned to mapping the SL regionspace to one theoretical quadrant of the
mapspace to avoid negative coordinates.


## Zoom levels ##

SL Maps zoom levels (Zl) start zoomed in, at Zl 1, where each tile comprises 1
region by itself. At Zl 2 each tile is 2x2 regions, and doubled again (in each
dimension) at each additional level. This means each tile is 2^(Zl-1) regions
across.

Leaflet lets us just use these zoom levels, though we have to specify our
levels are backwards & start at 1. The wrinkle with Leaflet is it wants to
address the tile images at different zoom levels in how many _tiles_ from (0,
0) we are, but SL map tiles are addressed in regionspace coordinates at all
zoom levels. These are the same at Zl 1 where 1 tile = 1 region. We provide the
`SLTileLayer` layer class to un-convert Leaflet's tile coordinates back to
regionspace coordinates at the other zoom levels.

Zoom levels in the Google Maps API were not only backwards, but organized such
that at Google Zoom (Zg) level 0 the entire worldspace was one (square) tile.
We needed a complex scheme for converting these at other steps too.

*/


// === Constants ===
var slDebugMap = $.sl.maps.config.map_debug;

var MIN_ZOOM_LEVEL = 1;
var MAX_ZOOM_LEVEL = 8;


/**
 * Creates a Second Life map in the given DOM element & returns the Leaflet Map.
 *
 * @param {Element} map_element the DOM element to contain the map
 */
function SLMap(mapElement, mapOptions)
{
    mapElement.className += ' slmapapi2-map-container';

    var mapDiv = document.createElement("div");
    mapDiv.style.height = "100%";
    mapElement.appendChild(mapDiv);

    var SLTileLayer = L.TileLayer.extend({
        getTileUrl: function (coords) {
            var data = {
                r: L.Browser.retina ? '@2x' : '',
                s: this._getSubdomain(coords),
                z: this._getZoomForUrl()
            };

            var regionsPerTileEdge = Math.pow(2, data['z'] - 1);
            data['region_x'] = coords.x * regionsPerTileEdge;
            data['region_y'] = (Math.abs(coords.y) - 1) * regionsPerTileEdge;

            return L.Util.template(this._url, L.extend(data, this.options));
        }
    });
    var tiles = new SLTileLayer($.sl.maps.config.tile_url + "/map-{z}-{region_x}-{region_y}-objects.jpg", {
        crs: L.CRS.Simple,
        minZoom: MIN_ZOOM_LEVEL,
        maxZoom: MAX_ZOOM_LEVEL,
        zoomOffset: 1,
        zoomReverse: true,
        bounds: [[0, 0], [$.sl.maps.config.map_grid_edge_size, $.sl.maps.config.map_grid_edge_size]],
        attribution: "<a href='" + $.sl.maps.config.sl_base_url + "'>Second Life</a>"
    });

    var map = L.map(mapDiv, {
        crs: L.CRS.Simple,
        minZoom: MIN_ZOOM_LEVEL,
        maxZoom: MAX_ZOOM_LEVEL,
        maxBounds: [[0, 0], [$.sl.maps.config.map_grid_edge_size, $.sl.maps.config.map_grid_edge_size]],
        layers: [tiles]
    });

    map.on('click', function (event) {
        gotoSLURL(event.latlng.lng, event.latlng.lat, map);
    });

    return map;
}

/**
 * Loads the script with the given URL by adding a script tag to the document.
 *
 * @private
 * @param {string} scriptURL the script to load
 * @param {function} onLoadHandler a callback to call when the script is loaded (optional)
 */
function slAddDynamicScript(scriptURL, onLoadHandler)
{
    var script = document.createElement('script');
    script.src = scriptURL;
    script.type = "text/javascript";

    if (onLoadHandler) {
        // Need to use ready state change for IE as it doesn't support onload for scripts
        script.onreadystatechange = function () {
            if (script.readyState == 'complete' || script.readyState == 'loaded') {
                onLoadHandler();
            }
        }

        // Standard onload for Firefox/Safari/Opera etc
        script.onload = onLoadHandler;
    }

    document.body.appendChild(script);
}

/**
 * Opens a map window (info window) for the given SL location, giving its name
 * and a "Teleport Here" button, as when clicked.
 *
 * @param {number} x the horizontal (west-east) SL Maps region coordinate to open the map window at
 * @param {number} y the vertical (south-north) SL Maps region coordinate to open the map window at
 * @param {SLMap} slMap the map in which to open the map window
 */
function gotoSLURL(x, y, lmap)
{
    // Work out region co-ords, and local co-ords within region
    var int_x = Math.floor(x);
    var int_y = Math.floor(y);

    var local_x = Math.round((x - int_x) * 256);
    var local_y = Math.round((y - int_y) * 256);

    // Add a dynamic script to get this region name, and then trigger a URL change
    // based on the results
    var scriptURL = "https://cap.secondlife.com/cap/0/b713fe80-283b-4585-af4d-a3b7d9a32492"
                    + "?var=slRegionName&grid_x=" + int_x + "&grid_y="+ int_y;

    // Once the script has loaded, we use the result to teleport the user into SL
    var onLoadHandler = function () {
        if (slRegionName == null || slRegionName.error)
            return;

        var url = "secondlife://" + encodeURIComponent(slRegionName)
                  + "/" + local_x + "/" + local_y;

        var debugInfo = '';
        if (slDebugMap) {
            debugInfo = ' x: ' + int_x + ' y: ' + int_y;
        }

        var content = '<div class="balloon-content balloon-content-narrow"><h3><a href="' + url + '">'+ slRegionName + '</a></h3>'
            + debugInfo
            + '<div class="buttons"><a href="'+ url +'" class="HIGHLANDER_button_hot btn_large primary">Visit this location</a></div></div>';
        var popup = L.popup().setLatLng([y, x]).setContent(content).openOn(lmap);
    };

    slAddDynamicScript(scriptURL, onLoadHandler);
}
;
/* ============================
** Build a SLurl validation
** ============================
** Makes sure there are slurl coordinates input before we attempt to generate a SLurl */


function slurlBuildValidator()
{
	$('input#generate-slurl').click(function(){
		
		$('#build-location div.slurl-error').remove(); //remove the error message if there is one
		
		var intX = parseFloat($('#x').val());
		var intY = parseFloat($('#y').val());
		var intZ = parseFloat($('#z').val());
		
		var imgRegExp = /(ftp|http|https):\/\/(\w+:{0,1}\w*@)?(\S+)(:[0-9]+)?(\/|\/([\w#!:.?+=&%@!\-\/]))?/;
		if($('#windowImage').val())	{
			var imgCheck = $('#windowImage').val().search(imgRegExp);
		}
		else {
			var imgCheck = 0;
		}

		
		if(!$('#region').val() 
			|| ( isNaN($('#x').val()) || isNaN(intX) || intX<0 || intX>256 )
			|| ( isNaN($('#y').val()) || isNaN(intY) || intY<0 || intY>256 )
			|| ( isNaN($('#z').val()) || isNaN(intZ) )
			|| ( $('#windowImage').val() != '' && imgCheck == -1 )
			) {
			
			var error_content = "";
			
			if(	!$('#region').val() ) {
				error_content += "<p>A valid SLurl must contain a region name and x,y,z coordinates. Please make sure that each of these fields are properly filled in.</p>";
			}
			
			if( isNaN($('#x').val()) || isNaN($('#y').val()) || isNaN($('#z').val()) || isNaN(intX) || isNaN(intY) || isNaN(intZ) || intX > 256 || intY > 256) {
				error_content += "<p>The x and y coordinates must each contain numbers between 0 and 256 to be valid. z index must be between -99 and 999. All coordinates (x,y,z) must be numeric.</p>";
			}
			
			if( intX<0 || intY<0) {
				error_content += "<p>The x and y coordinates must be positive numbers.</p>";
			}			
			
			if(imgCheck == -1) {
					error_content += "<p>Window Image must be a fully qualified url to a hosted image.";
					error_content += "( <strong>Example: </strong><em>http://www.yoursite.com/images/yourImage.jpg</em> )</p>";
			}
			
			$('#build-location legend').after('<div class="slurl-error"><h4>We\'re having trouble creating your SLurl</h4>'+ error_content +'</div>');
			
			document.getElementById('slurl-builder').scrollTop=0;
		} else {
			build_url();
		}
		
		
	})
}


/* ============================
** Build a SLurl
** ============================
** Creates the SLurl */

function build_url()
{
	var slurl = $('#slurl_base').val() + escape($('#region').val()) + "/" + parseFloat($('input#x').val()) + "/" + parseFloat($('input#y').val()) + "/" + parseFloat($('input#z').val()) + "/";

	$('#slurl-builder form #return-slurl').css({'display' : 'block'});

	document.getElementById('slurl-builder').scrollTop = $('#slurl-builder').height() + $('#return-slurl').height();

	// return the slurl to the output field
	$('#output').val(slurl);
	
	// fade out a color animation - requires color animation plugin
	$('#return-slurl').animate({ backgroundColor: '#f8f8f8' }, 2500);
	
}


/* =====================
** Popup windows
** =====================
** Creates an unobtrusive popup window */

function popUpExample()
{
	$('a.popup').click(function(){
		var href = $(this).attr('href');
		var width = screen.width - 760;
		var height = screen.height - 550;
		var left = (screen.width - width)/2;
		var top = (screen.height - height)/2;

		window.open(href, 'popup', 'height='+ height +',width='+ width +',left='+ left +',top='+ top +',toolbar=no,scrollbars=yes');

		return false;		
	});
}
;
// Find a url param by name
function gup( name,url )
{
    var regexS = "[\\?&]"+name+"=([^&#]*)";
    var regex = new RegExp( regexS );
    var tmpURL = (url === undefined) ? window.location.href : url;
    var results = regex.exec( tmpURL );
    if( results == null )
        return "";
    else
        return results[1];
}


function mapExtensions()
{
    // New jquery styling for drop-down menus
    //$('#search_select').dropp();

    // Open & close the sidebar.
    $('#map-search-results').click(function (evt) {
        var isCollapsed = $(this).hasClass('collapsed');
        var collapserClicked = $(evt.target).parents().addBack().filter('#collapse-new').length;
        if (!isCollapsed && !collapserClicked) {
            return;
        }

        $(this).toggleClass('collapsed');
        $('#map-container').toggleClass('map-search-opened');
    });
}


//For Google Analytics - creates a custom trackable item
function trakkit(category, action, opt_label, opt_value) {
    try { // depends on google analytics
        _gaq.push(['_trackEvent', category, action, opt_label]);
        //console.log('_trackEvent: '+category+','+action+','+opt_label);
    }
    catch(e) {
        //console.log('_trackEvent FAIL: '+e);
    }
}


/* ==============================================
** directSLurl
** ==============================================
** get a directSlurl on click of a login link */
function directSlurl($self, location_string)
{
    $.ajax({
        url: "/direct_slurl.php",
        type: "GET",
        data: { region_location_string: location_string },
        async: false,
        success: function(data) {
            if (data) $self.attr("href", data);
        }
    });
}


/* =========================================================================
**  loadquery
** =========================================================================
** ajaxish loading of search and showcase results,
** so we can load the map first */
function loadquery(doc, params)
{
    var error_response = "<div class=\"notice\" style=\"margin:60px 15px;\""
        +"<h3>We've encountered a problem</h3>"
        +"<p>We were unable to complete the search for the content you requested."
        +" Please continue exploring the map while we look into the cause. Thank you for your patience!</p>"
        +"</div>";
    var leftJustifyTitle = false;
    var ajaxParams = (leftJustifyTitle) ? "null" : params;
    $.ajax({
        url: doc,
        type: 'GET',
        data: ajaxParams,
        timeout: 10000,
        error: function(){
            $('#map-search-results .loader').remove();
            $('#map-search-results').append(error_response);
            mapquery();
        },
        success: function(data){
            $('#map-search-results .loader').remove();
            $('#map-search-results').append(data);

            /* FORMATTING THE TITLES AT THE TOP OF THE SIDEBAR */
            // If a destination is loaded, we adjust the CSS for the Destination Guide title
            if (leftJustifyTitle)
                $('#dest-guide-title').css('margin-left', '15px');

            // If a false destination is loaded, 'Your Location' is removed in a script in main.php.
            // Now we check to see if that happened, and if it did, we remove the duplicate title.
            if($('#location-heading').html() == 'Destination Guide Picks')
                $('#dest-guide-title').remove();
            /* END FORMATTING */

            // Make this results area of the sidebar scroll by limiting its
            // height. We don't know until it's added to the page how far down
            // the sidebar it starts (that is, how much stuff is above it in
            // the sidebar). So it has to be layed out as a relative element
            // without placement at first for us to find its top.
            var showcaseContainer = $('.showcase-container');
            if (showcaseContainer.length) {
                var showcaseContainerTop = showcaseContainer.position().top || '0px';
                showcaseContainer.css({
                    'position': 'absolute',
                    'top': showcaseContainerTop,
                    'bottom': '10px'
                });
            }

            mapquery();
        }
    });
}


/**
 * Makes a marker & map window tied to a particular sidebar search result or
 * Destination Guide item. When the marker is clicked & the map window opens,
 * the sidebar item becomes selected; it becomes deselected again when the map
 * window is closed. Clicking the sidebar item pans the map to the marker &
 * opens the map window.
 *
 * @param {Element|String} windowContent the content of the marker's map window
 * @param {Point} markerLocation the map location to place the marker at
 * @param {JQueryResult} $sidebarItem the jQuery set representing the marker's related sidebar item
 * @param {String} regionLocation the SL region location string to link the window's Join button to when the window is opened (optional)
 */
function makeMarkerForSidebar(windowContent, markerLatLng, $sidebarItem, regionLocation)
{
    var domId = $sidebarItem.attr('id');
    var marker = L.marker(markerLatLng, {
        icon: L.icon({
            iconUrl: assetsURL + "marker-sm.png",
            iconSize: [53, 48],
            iconAnchor: [26, 48]
        })
    }).addTo($map);
    // TODO: 350 is the declared width of the .balloon-content div wrapping the
    // windowContent. Someday we should figure out how to let Leaflet figure
    // that out instead of hardcoding it here?
    marker.bindPopup(windowContent, {minWidth: 350}).on('popupopen', function (event) {
        $('#' + domId).addClass('result-selected');

        if (!regionLocation) {
            return;
        }
        var $joinButton = $(event.popup.getPane()).find('a.join_button');
        if (0 < $joinButton.length) {
            directSlurl($joinButton, regionLocation);
        }
    }).on('popupclose', function (event) {
        $('#' + domId).removeClass('result-selected');
    });

    $sidebarItem.click(function (evt) {
        evt.preventDefault();
        marker.togglePopup();  // ???
    });

    return marker;
}


/* =========================================================================
**  loadmap
** =========================================================================
** A singular region or default region, no searches have been made */
function loadmap(firstJoinUrl)
{
    var zoomLevel = 6;

    // Select our old Ahern default, even if we don't put a marker there. This can
    // happen if the directly linked region name is no longer valid (the error
    // case), or until the sidebar loads.
    // var latlng = L.latLng(1002.5, 997.5);
	var latlng = L.latLng(3650.5, 3650.5);

    var $body = $('body');
    if ($body.data('region-coords-error')) {
        var errorContent = slurl_data['noexist_windowcontent'];

        // let's not link to nothing, now
        $('#marker0').remove();
        $('#location-heading').html('Destination Guide Picks');
        $('#dest-guide-title').remove();
        $('body').prepend(errorContent);

        $('#map-error #error-content .error-close').click(function () {
            $('#map-error').hide();
        });
        $map.setView(latlng, zoomLevel);
    }
    else if (!slurl_data['region']['default']) {
        var x = parseInt($body.data('region-coords-x')) + slurl_data['region']['x'] / 256;
        var y = parseInt($body.data('region-coords-y')) + slurl_data['region']['y'] / 256;
        if (is_number(x) && is_number(y)) latlng = L.latLng(y, x);

        var bubbleContent = slurl_data['windowcontent'].replace('/\s+/', ' ');
        var $content = $('<div/>').html(bubbleContent);
        if (firstJoinUrl) {
            $content.find('a.join_button').attr('href', firstJoinUrl);
        }

        var marker = makeMarkerForSidebar($content.get(0), latlng, $('#marker0'));
        $map.setView(latlng, MAX_ZOOM_LEVEL);
        marker.openPopup();
    }
    else {
        $map.setView(latlng, zoomLevel);
    }
}


/**
 * Try to decode the given string from a URI component to plain unencoded text.
 * In addition to the normal `decodeURIComponent()` decoding, plus `+`
 * characters, which are a historic URL encoding of spaces (such as by PHP
 * `urlencode`, vs `rawurlencode`), are decoded to spaces. If the given text
 * cannot be properly decoded (an unencoded percent sign `%` is present, the
 * encoded characters represent an invalid byte sequence in the current
 * character encoding, etc), the empty string is returned rather than a
 * JavaScript error being raised.
 *
 * @param {String} text the URI component to decode
 * @return {String} the decoded text represented by the URI component (or the empty string)
 */
function safeDecodeURIComponent(text)
{
    try {
        return decodeURIComponent(text.replace(/\+/g, ' '));
    }
    catch (err) {}
    return '';
}


/* =========================================================================
**  mapquery
** =========================================================================
** A query has been made, plot those points */
function mapquery()
{
    $("a.slurl-link").each(function (index) {
        var xy = $(this).attr("title").split(",");
        var rx = parseFloat(xy[0]);
        var ry = parseFloat(xy[1]);
        var linkHref = $(this).attr("href");

        var title = '';
        var msg = '';
        var img = '';

        /*//////////////////////////////////////////////////////////////////////
        // Must string replace the base url for IE because ie passes fully qualified url to javascript
        // even though it doesn't exist in the markup
        //////////////////////////////////////////////////////////////////////*/

        var re_tp_url = new RegExp('^(https?://[^/]+)?/secondlife/([^?]+).*$', 'i');
        // linkHref is still URI encoded, so tp_url ends up with the correct URI encoded name.
        var tp_region = linkHref.replace(re_tp_url, "$2");
        var tp_url = "secondlife://" + tp_region;

        if (rx && ry && tp_region) {
            var tokens = tp_region.split('/');
            loc_x = rx + parseFloat(tokens[1]) / 256;
            loc_y = ry + parseFloat(tokens[2]) / 256;
        }
        else {
            // Error with precise location plotting in the region - set to center
            loc_x = 128;
            loc_y = 128;
            msg = 'We\'re having difficulty plotting this location. It could be that this isn\'t a valid location, or that you are trying to locate a person.';
            tp_url = null;
        }

        var templateData = {
            'slurl': tp_url,
            'title': title || $.sl.maps.config.default_title,
            'img': img || $.sl.maps.config.default_img,
            'msg': msg || $.sl.maps.config.default_msg
        };

        Mustache.parse($.sl.maps.config.balloon_tmpl);
        var windowContent = Mustache.render($.sl.maps.config.balloon_tmpl, templateData);
        var markerLatLng = L.latLng(loc_y, loc_x);
        var $result = $(this).parents('div.result').first();
        var regionLocation = '/secondlife/' + tp_region;

        makeMarkerForSidebar(windowContent, markerLatLng, $result, regionLocation);
    });

    // Select the first result if we didn't load with a place selected (that
    // is, a search or a default maps.sl.com page load).
    if (slurl_data['region']['default']) {
        $("a.slurl-link").first().click();
    }
}
;
var urlParams = {};
var slurl_data = new Object();

(function () {
    var e,
        a = /\+/g,  // Regex for replacing addition symbol with a space
        r = /([^&=]+)=?([^&]*)/g,
        d = function (s) { return decodeURIComponent(s.replace(a, " ")); },
        q = window.location.search.substring(1);

    while (e = r.exec(q))
       urlParams[d(e[1])] = safeDecodeURIComponent(e[2]);
})();

// Aggregates all data pertaining to a slurl - Includes title, image, description, x/y coordinates, etc
function slurl_setup() {
    var urlPath = window.location.pathname;
    var urlParts = urlPath.split("/");
 //   var initial_region = 'Ahern';
 	var initial_region = 'Beta Technologies';
    slurl_data['region'] = new Object();

    if (urlParts[1] == '' || urlParts[1] == 'index.php' || urlParams['q'] != undefined || urlParts[3] == undefined) {
        slurl_data['region']['name'] = initial_region;
        slurl_data['region']['x'] = 0;
        slurl_data['region']['y'] = 0;
        slurl_data['region']['z'] = 0;
        slurl_data['region']['default'] = true;
    }
    else
    {
        initial_region = (urlParts[2] == undefined) ? 'Beta Technologies' : decodeURIComponent(urlParts[2]);
        slurl_data['region']['name'] = initial_region;
        // The following 2 should be numeric only -- need to regex 0-256
        slurl_data['region']['x'] = (urlParts[3] != undefined) ? check_coords(urlParts[3]) : 128 ;
        slurl_data['region']['y'] = (urlParts[4] != undefined) ? check_coords(urlParts[4]) : 128 ;
        // z-index can much higher than 256, 4 digits at least, here we have an uncapped check
        slurl_data['region']['z'] = (urlParts[5] != undefined) ? urlParts[5] : 0;
        slurl_data['region']['default'] = false;
    }

    // hurl the slurl
    var slurl = 'secondlife://' + encodeURIComponent(slurl_data['region']['name'].toUpperCase()) + '/' + slurl_data['region']['x'] + '/' + slurl_data['region']['y'] + '/' + slurl_data['region']['z'];

    // Setup the template data
    var templateData = {
        'slurl': slurl,
        'title': $.sl.maps.config.default_title,
        'img': $.sl.maps.config.default_img,
        'msg': $.sl.maps.config.default_msg
    };

    // Create the balloon content
    Mustache.parse($.sl.maps.config.balloon_tmpl);
    slurl_data['windowcontent'] = Mustache.render($.sl.maps.config.balloon_tmpl, templateData);

    // Some content for a sidebar if needed
    slurl_data['sidebarcontent'] = templateData;

    // A default for when a region doesn't exist.
    Mustache.parse($.sl.maps.config.noexists_tmpl);
    slurl_data['noexist_windowcontent'] = Mustache.render($.sl.maps.config.noexists_tmpl, {'region_name':slurl_data['region']['name']} );
}

// Check that a coordinate within a region is valid.  If not, return the center of the region
function check_coords(slurl_coord)
{
    if (is_number(slurl_coord) && ((0 <= slurl_coord) && (slurl_coord <= 256))) {
        return slurl_coord;
    }
    else {
        return 128;
    }
}

function is_number(n) {
  return !isNaN(parseFloat(n)) && isFinite(n);
}

slurl_setup();
// application.js









;