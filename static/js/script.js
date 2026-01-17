const hamburger = document.querySelector('#hamburger');
const nav = document.querySelector('nav');
const main = document.querySelector('main');

hamburger.addEventListener('click', () => {
    hamburger.classList.toggle('open');
    nav.classList.toggle('open');
    main.classList.toggle('moved');

    document.body.classList.toggle("frozen");
});

document.addEventListener('click', (event) => {
    console.log('user clicked on:', event.target);
});
