<?xml version="1.0" encoding="UTF-8"?>
<!--
  Vedic Transit XSLT (Gochara)

  Consumes TransitChart XML. Computes:
  - Nakshatra gochara (transit planet nakshatra positions)
  - Sidereal whole-sign house overlays
  - Vimshottari dasha context (current dasha at transit moment)
  - Vedic dignities (swakshetra, uchcha, neecha)
-->
<xsl:stylesheet xmlns:xsl="http://www.w3.org/1999/XSL/Transform"
                xmlns:math="http://exslt.org/math"
                xmlns:exslt="http://exslt.org/common"
                xmlns:vedic="urn:empirical:vedic"
                version="1.0"
                exclude-result-prefixes="math exslt">

  <!-- ── Parameters ────────────────────────────────────────────────────── -->
  <xsl:param name="orb" select="3.0"/>

  <!-- ── Vedic planets (9 grahas) ───────────────────────────────────────── -->
  <xsl:variable name="grahas" select="',Sun,Moon,Mercury,Venus,Mars,Jupiter,Saturn,Rahu,Ketu,'"/>

  <!-- ── Sign names ────────────────────────────────────────────────────── -->
  <xsl:variable name="signs" select="'Aries,Taurus,Gemini,Cancer,Leo,Virgo,Libra,Scorpio,Sagittarius,Capricorn,Aquarius,Pisces'"/>

  <!-- ── Nakshatra names (27) ───────────────────────────────────────────── -->
  <xsl:variable name="nakshatras" select="'Ashwini,Bharani,Krittika,Rohini,Mrigashira,Ardra,Punarvasu,Pushya,Ashlesha,Magha,Purva Phalguni,Uttara Phalguni,Hasta,Chitra,Swati,Vishakha,Anuradha,Jyeshtha,Mula,Purva Ashadha,Uttara Ashadha,Shravana,Dhanishta,Shatabhisha,Purva Bhadrapada,Uttara Bhadrapada,Revati'"/>

  <!-- ── Nakshatra lords (Vimshottari sequence) ─────────────────────────── -->
  <xsl:variable name="nakshatraLords" select="'Ketu,Venus,Sun,Moon,Mars,Rahu,Jupiter,Saturn,Mercury'"/>

  <!-- ── Vimshottari dasha years ────────────────────────────────────────── -->
  <xsl:variable name="dashaYearsKetu" select="7"/>
  <xsl:variable name="dashaYearsVenus" select="20"/>
  <xsl:variable name="dashaYearsSun" select="6"/>
  <xsl:variable name="dashaYearsMoon" select="10"/>
  <xsl:variable name="dashaYearsMars" select="7"/>
  <xsl:variable name="dashaYearsRahu" select="18"/>
  <xsl:variable name="dashaYearsJupiter" select="16"/>
  <xsl:variable name="dashaYearsSaturn" select="19"/>
  <xsl:variable name="dashaYearsMercury" select="17"/>

  <!-- ── Domicile (swakshetra) ──────────────────────────────────────────── -->
  <xsl:variable name="swakshetra0" select="'Mars'"/>     <!-- Aries -->
  <xsl:variable name="swakshetra1" select="'Venus'"/>    <!-- Taurus -->
  <xsl:variable name="swakshetra2" select="'Mercury'"/>  <!-- Gemini -->
  <xsl:variable name="swakshetra3" select="'Moon'"/>     <!-- Cancer -->
  <xsl:variable name="swakshetra4" select="'Sun'"/>      <!-- Leo -->
  <xsl:variable name="swakshetra5" select="'Mercury'"/>  <!-- Virgo -->
  <xsl:variable name="swakshetra6" select="'Venus'"/>    <!-- Libra -->
  <xsl:variable name="swakshetra7" select="'Mars'"/>     <!-- Scorpio -->
  <xsl:variable name="swakshetra8" select="'Jupiter'"/>  <!-- Sagittarius -->
  <xsl:variable name="swakshetra9" select="'Saturn'"/>   <!-- Capricorn -->
  <xsl:variable name="swakshetra10" select="'Saturn'"/>  <!-- Aquarius -->
  <xsl:variable name="swakshetra11" select="'Jupiter'"/> <!-- Pisces -->

  <!-- ── Exaltation (uchcha) ────────────────────────────────────────────── -->
  <xsl:variable name="uchcha0" select="'Sun'"/>      <!-- Aries: Sun 10° -->
  <xsl:variable name="uchcha1" select="'Moon'"/>     <!-- Taurus: Moon 3° -->
  <xsl:variable name="uchcha2" select="'Rahu'"/>     <!-- Gemini: Rahu -->
  <xsl:variable name="uchcha3" select="'Jupiter'"/>  <!-- Cancer: Jupiter 5° -->
  <xsl:variable name="uchcha4" select="''"/>         <!-- Leo: none -->
  <xsl:variable name="uchcha5" select="'Mercury'"/>  <!-- Virgo: Mercury 15° -->
  <xsl:variable name="uchcha6" select="'Saturn'"/>   <!-- Libra: Saturn 20° -->
  <xsl:variable name="uchcha7" select="''"/>         <!-- Scorpio: none -->
  <xsl:variable name="uchcha8" select="'Ketu'"/>     <!-- Sagittarius: Ketu -->
  <xsl:variable name="uchcha9" select="'Mars'"/>     <!-- Capricorn: Mars 28° -->
  <xsl:variable name="uchcha10" select="''"/>        <!-- Aquarius: none -->
  <xsl:variable name="uchcha11" select="'Venus'"/>   <!-- Pisces: Venus 27° -->

  <!-- ── Debilitation (neecha) ──────────────────────────────────────────── -->
  <xsl:variable name="neecha0" select="'Saturn'"/>   <!-- Aries: Saturn -->
  <xsl:variable name="neecha1" select="''"/>         <!-- Taurus: none -->
  <xsl:variable name="neecha2" select="'Ketu'"/>     <!-- Gemini: Ketu -->
  <xsl:variable name="neecha3" select="'Mars'"/>     <!-- Cancer: Mars -->
  <xsl:variable name="neecha4" select="''"/>         <!-- Leo: none -->
  <xsl:variable name="neecha5" select="'Venus'"/>    <!-- Virgo: Venus -->
  <xsl:variable name="neecha6" select="'Sun'"/>      <!-- Libra: Sun -->
  <xsl:variable name="neecha7" select="'Moon'"/>     <!-- Scorpio: Moon -->
  <xsl:variable name="neecha8" select="'Rahu'"/>    <!-- Sagittarius: Rahu -->
  <xsl:variable name="neecha9" select="'Jupiter'"/>  <!-- Capricorn: Jupiter -->
  <xsl:variable name="neecha10" select="''"/>        <!-- Aquarius: none -->
  <xsl:variable name="neecha11" select="'Mercury'"/> <!-- Pisces: Mercury -->

  <!-- ══════════════════════════════════════════════════════════════════════
       ROOT TEMPLATE
       ════════════════════════════════════════════════════════════════════ -->
  <xsl:template match="/TransitChart">
    <vedic:TransitReport xmlns:vedic="urn:empirical:vedic">
      <vedic:Name><xsl:value-of select="Identity/Name"/></vedic:Name>
      <vedic:TransitDate>
        <xsl:value-of select="Moment/Year"/>-<xsl:value-of select="format-number(Moment/Month,'00')"/>-<xsl:value-of select="format-number(Moment/Day,'00')"/>
      </vedic:TransitDate>

      <!-- Capture natal sidereal ASC for whole-sign house overlay -->
      <xsl:variable name="natalSidASC" select="(Natal/Angles/ASC - Natal/Positions/Ayanamsa + 360) mod 360"/>

      <!-- ══════════════════════════════════════════════════════════════════
           NAKSHATRA GOCHARA (transit planet nakshatra positions)
           ════════════════════════════════════════════════════════════════ -->
      <vedic:Gochara>
        <xsl:for-each select="Transits/Planet[contains($grahas, concat(',',@name,','))]">
          <xsl:variable name="planet" select="@name"/>
          <!-- Use sidereal longitude -->
          <xsl:variable name="sidLon" select="Sidereal/Lon"/>
          <xsl:variable name="nakshatraIdx" select="floor($sidLon div 13.333333)"/>
          <xsl:variable name="pada" select="floor(($sidLon mod 13.333333) div 3.333333) + 1"/>
          <xsl:variable name="sign" select="floor($sidLon div 30)"/>

          <vedic:PlanetGochara>
            <vedic:Planet><xsl:value-of select="$planet"/></vedic:Planet>
            <vedic:Sign>
              <xsl:call-template name="signName">
                <xsl:with-param name="idx" select="$sign"/>
              </xsl:call-template>
            </vedic:Sign>
            <vedic:Degree><xsl:value-of select="format-number($sidLon mod 30,'0.00')"/></vedic:Degree>
            <vedic:Nakshatra>
              <xsl:call-template name="nthToken">
                <xsl:with-param name="list" select="$nakshatras"/>
                <xsl:with-param name="n" select="$nakshatraIdx + 1"/>
                <xsl:with-param name="delimiter" select="','"/>
              </xsl:call-template>
            </vedic:Nakshatra>
            <vedic:Pada><xsl:value-of select="$pada"/></vedic:Pada>
            <vedic:NakshatraLord>
              <xsl:call-template name="nthToken">
                <xsl:with-param name="list" select="$nakshatraLords"/>
                <xsl:with-param name="n" select="($nakshatraIdx mod 9) + 1"/>
                <xsl:with-param name="delimiter" select="','"/>
              </xsl:call-template>
            </vedic:NakshatraLord>
            <!-- Whole-sign house overlay (sidereal) -->
            <vedic:NatalHouse>
              <xsl:call-template name="wholeSignHouse">
                <xsl:with-param name="lon" select="$sidLon"/>
                <xsl:with-param name="asc" select="$natalSidASC"/>
              </xsl:call-template>
            </vedic:NatalHouse>
          </vedic:PlanetGochara>
        </xsl:for-each>
      </vedic:Gochara>

      <!-- ══════════════════════════════════════════════════════════════════
           TRANSIT DIGNITIES (Vedic)
           ════════════════════════════════════════════════════════════════ -->
      <vedic:TransitDignities>
        <xsl:for-each select="Transits/Planet[contains($grahas, concat(',',@name,','))]">
          <xsl:variable name="planet" select="@name"/>
          <xsl:variable name="sidLon" select="Sidereal/Lon"/>
          <xsl:variable name="sign" select="floor($sidLon div 30)"/>

          <vedic:TransitDignity>
            <vedic:Planet><xsl:value-of select="$planet"/></vedic:Planet>
            <vedic:Sign>
              <xsl:call-template name="signName">
                <xsl:with-param name="idx" select="$sign"/>
              </xsl:call-template>
            </vedic:Sign>
            <vedic:Swakshetra>
              <xsl:call-template name="checkSwakshetra">
                <xsl:with-param name="planet" select="$planet"/>
                <xsl:with-param name="sign" select="$sign"/>
              </xsl:call-template>
            </vedic:Swakshetra>
            <vedic:Uchcha>
              <xsl:call-template name="checkUchcha">
                <xsl:with-param name="planet" select="$planet"/>
                <xsl:with-param name="sign" select="$sign"/>
              </xsl:call-template>
            </vedic:Uchcha>
            <vedic:Neecha>
              <xsl:call-template name="checkNeecha">
                <xsl:with-param name="planet" select="$planet"/>
                <xsl:with-param name="sign" select="$sign"/>
              </xsl:call-template>
            </vedic:Neecha>
          </vedic:TransitDignity>
        </xsl:for-each>
      </vedic:TransitDignities>

    </vedic:TransitReport>
  </xsl:template>

  <!-- ══════════════════════════════════════════════════════════════════════
       HELPER TEMPLATES
       ════════════════════════════════════════════════════════════════════ -->

  <xsl:template name="signName">
    <xsl:param name="idx"/>
    <xsl:call-template name="nthToken">
      <xsl:with-param name="list" select="$signs"/>
      <xsl:with-param name="n" select="$idx + 1"/>
    </xsl:call-template>
  </xsl:template>

  <xsl:template name="wholeSignHouse">
    <xsl:param name="lon"/>
    <xsl:param name="asc"/>
    <xsl:variable name="sign" select="floor($lon div 30)"/>
    <xsl:variable name="ascSign" select="floor($asc div 30)"/>
    <xsl:variable name="house" select="(($sign - $ascSign + 12) mod 12) + 1"/>
    <xsl:value-of select="$house"/>
  </xsl:template>

  <xsl:template name="checkSwakshetra">
    <xsl:param name="planet"/>
    <xsl:param name="sign"/>
    <xsl:variable name="ruler">
      <xsl:choose>
        <xsl:when test="$sign = 0"><xsl:value-of select="$swakshetra0"/></xsl:when>
        <xsl:when test="$sign = 1"><xsl:value-of select="$swakshetra1"/></xsl:when>
        <xsl:when test="$sign = 2"><xsl:value-of select="$swakshetra2"/></xsl:when>
        <xsl:when test="$sign = 3"><xsl:value-of select="$swakshetra3"/></xsl:when>
        <xsl:when test="$sign = 4"><xsl:value-of select="$swakshetra4"/></xsl:when>
        <xsl:when test="$sign = 5"><xsl:value-of select="$swakshetra5"/></xsl:when>
        <xsl:when test="$sign = 6"><xsl:value-of select="$swakshetra6"/></xsl:when>
        <xsl:when test="$sign = 7"><xsl:value-of select="$swakshetra7"/></xsl:when>
        <xsl:when test="$sign = 8"><xsl:value-of select="$swakshetra8"/></xsl:when>
        <xsl:when test="$sign = 9"><xsl:value-of select="$swakshetra9"/></xsl:when>
        <xsl:when test="$sign = 10"><xsl:value-of select="$swakshetra10"/></xsl:when>
        <xsl:when test="$sign = 11"><xsl:value-of select="$swakshetra11"/></xsl:when>
      </xsl:choose>
    </xsl:variable>
    <xsl:choose>
      <xsl:when test="$planet = $ruler">true</xsl:when>
      <xsl:otherwise>false</xsl:otherwise>
    </xsl:choose>
  </xsl:template>

  <xsl:template name="checkUchcha">
    <xsl:param name="planet"/>
    <xsl:param name="sign"/>
    <xsl:variable name="ruler">
      <xsl:choose>
        <xsl:when test="$sign = 0"><xsl:value-of select="$uchcha0"/></xsl:when>
        <xsl:when test="$sign = 1"><xsl:value-of select="$uchcha1"/></xsl:when>
        <xsl:when test="$sign = 2"><xsl:value-of select="$uchcha2"/></xsl:when>
        <xsl:when test="$sign = 3"><xsl:value-of select="$uchcha3"/></xsl:when>
        <xsl:when test="$sign = 4"><xsl:value-of select="$uchcha4"/></xsl:when>
        <xsl:when test="$sign = 5"><xsl:value-of select="$uchcha5"/></xsl:when>
        <xsl:when test="$sign = 6"><xsl:value-of select="$uchcha6"/></xsl:when>
        <xsl:when test="$sign = 7"><xsl:value-of select="$uchcha7"/></xsl:when>
        <xsl:when test="$sign = 8"><xsl:value-of select="$uchcha8"/></xsl:when>
        <xsl:when test="$sign = 9"><xsl:value-of select="$uchcha9"/></xsl:when>
        <xsl:when test="$sign = 10"><xsl:value-of select="$uchcha10"/></xsl:when>
        <xsl:when test="$sign = 11"><xsl:value-of select="$uchcha11"/></xsl:when>
      </xsl:choose>
    </xsl:variable>
    <xsl:choose>
      <xsl:when test="$planet = $ruler">true</xsl:when>
      <xsl:otherwise>false</xsl:otherwise>
    </xsl:choose>
  </xsl:template>

  <xsl:template name="checkNeecha">
    <xsl:param name="planet"/>
    <xsl:param name="sign"/>
    <xsl:variable name="ruler">
      <xsl:choose>
        <xsl:when test="$sign = 0"><xsl:value-of select="$neecha0"/></xsl:when>
        <xsl:when test="$sign = 1"><xsl:value-of select="$neecha1"/></xsl:when>
        <xsl:when test="$sign = 2"><xsl:value-of select="$neecha2"/></xsl:when>
        <xsl:when test="$sign = 3"><xsl:value-of select="$neecha3"/></xsl:when>
        <xsl:when test="$sign = 4"><xsl:value-of select="$neecha4"/></xsl:when>
        <xsl:when test="$sign = 5"><xsl:value-of select="$neecha5"/></xsl:when>
        <xsl:when test="$sign = 6"><xsl:value-of select="$neecha6"/></xsl:when>
        <xsl:when test="$sign = 7"><xsl:value-of select="$neecha7"/></xsl:when>
        <xsl:when test="$sign = 8"><xsl:value-of select="$neecha8"/></xsl:when>
        <xsl:when test="$sign = 9"><xsl:value-of select="$neecha9"/></xsl:when>
        <xsl:when test="$sign = 10"><xsl:value-of select="$neecha10"/></xsl:when>
        <xsl:when test="$sign = 11"><xsl:value-of select="$neecha11"/></xsl:when>
      </xsl:choose>
    </xsl:variable>
    <xsl:choose>
      <xsl:when test="$planet = $ruler">true</xsl:when>
      <xsl:otherwise>false</xsl:otherwise>
    </xsl:choose>
  </xsl:template>

  <xsl:template name="nthToken">
    <xsl:param name="list"/>
    <xsl:param name="n"/>
    <xsl:param name="delimiter" select="','"/>
    <xsl:choose>
      <xsl:when test="$n = 1">
        <xsl:choose>
          <xsl:when test="contains($list, $delimiter)">
            <xsl:value-of select="substring-before($list, $delimiter)"/>
          </xsl:when>
          <xsl:otherwise>
            <xsl:value-of select="$list"/>
          </xsl:otherwise>
        </xsl:choose>
      </xsl:when>
      <xsl:when test="contains($list, $delimiter)">
        <xsl:call-template name="nthToken">
          <xsl:with-param name="list" select="substring-after($list, $delimiter)"/>
          <xsl:with-param name="n" select="$n - 1"/>
          <xsl:with-param name="delimiter" select="$delimiter"/>
        </xsl:call-template>
      </xsl:when>
    </xsl:choose>
  </xsl:template>

</xsl:stylesheet>
