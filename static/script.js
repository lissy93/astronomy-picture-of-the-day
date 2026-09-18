/** Fetches today's APOD from the backend and renders it into the page. */

/* API endpoints, using the current origin or the public instance as a fallback. */
const endpoints = () => {
  const origin = window.location.origin;
  const base = origin && origin !== 'null' ? origin : 'https://apod.as93.net';
  return { apod: `${base}/apod`, image: `${base}/image` };
};

const byId = (id) => document.getElementById(id);
const hide = (el) => { if (el) el.style.display = 'none'; };
const show = (el) => { if (el) el.style.display = 'block'; };

/* Build the URL of the original APOD page for a date (YYYY-MM-DD). */
const nasaPageUrl = (date) => {
  const base = 'https://science.nasa.gov/apod/';
  return date ? `${base}?date=${encodeURIComponent(date)}` : base;
};

/* Format an ISO date as a human-readable, localised date. */
const formatDate = (date) => {
  if (!date) return '';
  return new Date(`${date}T00:00:00`).toLocaleDateString(undefined, {
    weekday: 'long', year: 'numeric', month: 'long', day: 'numeric',
  });
};

/* Point the HD link at a URL, or hide it if there is nothing to link to. */
const setLink = (href, text) => {
  const link = byId('apod-hd-link');
  if (!href) { hide(link); return; }
  link.href = href;
  link.innerText = text;
  show(link);
};

/* Render the APOD response into the DOM, handling each media type. */
const render = (apod) => {
  show(byId('apod-info'));
  byId('apod-title').innerText = apod.title || 'Astronomy Picture of the Day';
  byId('apod-explanation').innerText = apod.explanation || '';
  byId('apod-copyright').innerText = apod.copyright || '';
  byId('apod-date').innerText = formatDate(apod.date);
  byId('response').innerHTML = prettyPrint(apod);

  const image = byId('apod-picture');
  const frame = byId('apod-dynamic-content');
  hide(frame); // only shown for video days

  if (apod.media_type === 'video' && apod.url) {
    hide(image);
    frame.src = apod.url;
    frame.title = apod.title || 'Astronomy Picture of the Day';
    show(frame);
    setLink(apod.url, 'Watch Video');
  } else if (apod.url || apod.hdurl) {
    setLink(apod.hdurl || apod.url, 'View HD Image');
  } else {
    // media_type "other": nothing to embed, so link to the source page.
    hide(image);
    setLink(nasaPageUrl(apod.date), 'View on NASA');
  }
};

/* Show the error box and log the detail. */
const showError = (err) => {
  console.error(err);
  hide(byId('plate'));
  show(byId('error'));
};

/* Syntax-highlight the JSON response shown in the API docs. */
const prettyPrint = (json) => {
  if (typeof json !== 'string') { json = JSON.stringify(json, null, 2); }
  json = json.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;');
  return json.replace(/("(\\u[a-zA-Z0-9]{4}|\\[^u]|[^\\"])*"(\s*:)?|\b(true|false|null)\b|-?\d+(?:\.\d*)?(?:[eE][+\-]?\d+)?)/g, (match) => {
    let cls = 'number';
    if (/^"/.test(match)) { cls = /:$/.test(match) ? 'key' : 'string'; }
    else if (/true|false/.test(match)) { cls = 'boolean'; }
    else if (/null/.test(match)) { cls = 'null'; }
    return `<span class="${cls}">${match}</span>`;
  });
};

/* Fetch today's APOD and render it. */
const load = () => {
  fetch(endpoints().apod)
    .then((res) => {
      if (!res.ok) throw new Error(`APOD API responded with ${res.status}`);
      return res.json();
    })
    .then(render)
    .catch(showError)
    .finally(() => hide(byId('loader')));
};

/* Show the live endpoint URLs in the API docs. */
const setApiEndpoints = () => {
  const { apod, image } = endpoints();
  byId('get-apod').innerText = apod;
  byId('get-img').innerText = image;
};

document.addEventListener('DOMContentLoaded', () => {
  const image = byId('apod-picture');
  // Hide the eager <img src="/image"> if it fails (e.g. a media_type "other" day).
  image.addEventListener('error', () => hide(image));
  if (image.complete && image.naturalWidth === 0) hide(image);

  load();
  setApiEndpoints();
});
